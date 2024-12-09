package maddrproxy

import (
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/hrntknr/maddr-proxy/pkg/utils"
	"github.com/vishvananda/netlink"
)

const iprouteProtocol = 151
const priority = 15100

const tableMain = 254
const tableRangeStart = 15100
const tableRangeEnd = 15199

func SetupRoute(watch bool, iface []string) error {
	if watch {
		route := make(chan netlink.RouteUpdate)
		addr := make(chan netlink.AddrUpdate)
		link := make(chan netlink.LinkUpdate)
		if err := netlink.RouteSubscribe(route, nil); err != nil {
			return err
		}
		if err := netlink.AddrSubscribe(addr, nil); err != nil {
			return err
		}
		if err := netlink.LinkSubscribe(link, nil); err != nil {
			return err
		}
		for {
			if err := ensureSetupRoute(iface); err != nil {
				return err
			}
			select {
			case <-route:
			case <-addr:
			case <-link:
			}
		}
	} else {
		return ensureSetupRoute(iface)
	}
}

func ensureSetupRoute(iface []string) error {
	mapping, err := ensureRules(iface)
	if err != nil {
		return err
	}
	if err := ensureRoutes(mapping); err != nil {
		return err
	}
	return nil
}

func ensureRoutes(mapping map[int]map[int]int) error {
	for family, m := range mapping {
		for table, device := range m {
			if err := ensureRoute(family, table, device); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureRoute(family int, table int, device int) error {
	routes, err := getRoutes(family, table)
	if err != nil {
		return err
	}

	find := false
	for _, route := range routes {
		if isDefaultRoute(route.Dst) {
			find = true
		} else {
			if err := netlink.RouteDel(&route); err != nil {
				return fmt.Errorf("failed to delete route: %w", err)
			}
		}
	}
	if !find {
		if err := netlink.RouteAdd(&netlink.Route{
			Dst:       getDefaultRoute(family),
			LinkIndex: device,
			Scope:     netlink.SCOPE_UNIVERSE,
			Protocol:  iprouteProtocol,
			Table:     table,
		}); err != nil {
			return fmt.Errorf("failed to add route: %w", err)
		}
	}
	return nil
}

func getDefaultRoute(family int) *net.IPNet {
	switch family {
	case netlink.FAMILY_V4:
		return &net.IPNet{IP: net.IPv4zero, Mask: net.IPMask(net.IPv4zero)}
	case netlink.FAMILY_V6:
		return &net.IPNet{IP: net.IPv6zero, Mask: net.IPMask(net.IPv6zero)}
	}
	return nil
}

func ensureRules(iface []string) (map[int]map[int]int, error) {
	links, err := getLinks(iface)
	if err != nil {
		return nil, err
	}
	ret := map[int]map[int]int{}
	for _, family := range []int{netlink.FAMILY_V4, netlink.FAMILY_V6} {
		defaultRouteIface, err := getDefaultRouteIface(family)
		if err != nil {
			return nil, err
		}
		newLinks := []netlink.Link{}
		for _, link := range links {
			if link.Attrs().Index != defaultRouteIface {
				newLinks = append(newLinks, link)
			}
		}
		addrs, err := getAddrList(newLinks, family)
		if err != nil {
			return nil, err
		}
		mapping, err := ensureRule(addrs, family)
		if err != nil {
			return nil, err
		}
		ret[family] = mapping
	}
	return ret, nil
}

func getRoutes(family int, table int) ([]netlink.Route, error) {
	routes, err := netlink.RouteListFiltered(family, &netlink.Route{Table: table}, netlink.RT_FILTER_TABLE)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func getLinks(iface []string) ([]netlink.Link, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	filtered := []netlink.Link{}
	for _, link := range links {
		for _, i := range iface {
			if regexp.MustCompile(i).MatchString(link.Attrs().Name) {
				filtered = append(filtered, link)
				break
			}
		}
	}
	return filtered, nil
}

func getDefaultRouteIface(family int) (int, error) {
	routes, err := netlink.RouteList(nil, family)
	if err != nil {
		return 0, err
	}
	for _, route := range routes {
		if isDefaultRoute(route.Dst) && route.Table == tableMain {
			return route.LinkIndex, nil
		}
	}
	return 0, errors.New("no default route")
}

func getAddrList(links []netlink.Link, family int) ([]netlink.Addr, error) {
	addrs := []netlink.Addr{}
	for _, link := range links {
		_addrs, err := netlink.AddrList(link, family)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, _addrs...)
	}
	filtered := []netlink.Addr{}
	for _, addr := range addrs {
		switch family {
		case netlink.FAMILY_V4:
			if utils.IsValidIPv4(addr.IP) {
				filtered = append(filtered, addr)
			}
		case netlink.FAMILY_V6:
			if utils.IsValidIPv6(addr.IP) {
				filtered = append(filtered, addr)
			}
		}
	}
	return filtered, nil
}

func ensureRule(addrs []netlink.Addr, family int) (map[int]int, error) {
	rules, err := netlink.RuleList(family)
	if err != nil {
		return nil, err
	}
	filtered := []netlink.Rule{}
	for _, rule := range rules {
		if rule.Table >= tableRangeStart && rule.Table <= tableRangeEnd {
			filtered = append(filtered, rule)
		}
	}

	mapping := map[int]int{}

	for _, rule := range filtered {
		found := false
		for _, addr := range addrs {
			if isDefaultRoute(rule.Dst) && matchIP(family, rule.Src, addr.IP) {
				found = true
				break
			}
		}
		if !found {
			if err := netlink.RuleDel(&rule); err != nil {
				return nil, fmt.Errorf("failed to delete rule: %w", err)
			}
		}
	}
	for _, addr := range addrs {
		found := false
		for _, rule := range filtered {
			if isDefaultRoute(rule.Dst) && matchIP(family, rule.Src, addr.IP) {
				mapping[rule.Table] = addr.LinkIndex
				found = true
				break
			}
		}
		if !found {
			table, err := findTable(family)
			if err != nil {
				return nil, err
			}
			mask := getMaskSize(family)
			rule := netlink.NewRule()
			rule.Family = family
			rule.Priority = priority
			rule.Table = table
			rule.Src = &net.IPNet{IP: addr.IP, Mask: net.CIDRMask(mask, mask)}
			if err := netlink.RuleAdd(rule); err != nil {
				return nil, fmt.Errorf("failed to add rule: %w", err)
			}
			mapping[table] = addr.LinkIndex
		}
	}

	return mapping, nil
}

func matchIP(family int, target *net.IPNet, addr net.IP) bool {
	if target == nil {
		return false
	}
	size, _ := target.Mask.Size()
	if size != getMaskSize(family) {
		return false
	}
	return target.IP.Equal(addr)
}

func getMaskSize(family int) int {
	switch family {
	case netlink.FAMILY_V4:
		return 32
	case netlink.FAMILY_V6:
		return 128
	default:
		return 0
	}
}

func findTable(family int) (int, error) {
	rules, err := netlink.RuleList(family)
	if err != nil {
		return 0, err
	}
	used := map[int]struct{}{}
	for _, rule := range rules {
		used[rule.Table] = struct{}{}
	}
	for i := tableRangeStart; i <= tableRangeEnd; i++ {
		if _, ok := used[i]; !ok {
			return i, nil
		}
	}
	return 0, errors.New("no available table number")
}

func isDefaultRoute(ipnet *net.IPNet) bool {
	if ipnet == nil {
		return true
	}
	if size, _ := ipnet.Mask.Size(); size == 0 {
		return true
	}
	return false
}
