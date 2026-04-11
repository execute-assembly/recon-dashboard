package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ComputeTotalIPs(cidr string) (int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}
	ones, bits := ipNet.Mask.Size()
	return 1 << (bits - ones), nil
}

func FetchAsnData(asn string) (AsnResult, error) {
	urlOverview := fmt.Sprintf("https://stat.ripe.net/data/as-overview/data.json?resource=%s", asn)
	urlPrefixes := fmt.Sprintf("https://stat.ripe.net/data/announced-prefixes/data.json?resource=%s", asn)

	resp, err := http.Get(urlOverview)
	if err != nil {
		return AsnResult{}, err
	}
	defer resp.Body.Close()

	var overview RipeOverviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return AsnResult{}, err
	}

	resp2, err := http.Get(urlPrefixes)
	if err != nil {
		return AsnResult{}, err
	}
	defer resp2.Body.Close()

	var AsnPrefix RipePrefixResponse
	if err := json.NewDecoder(resp2.Body).Decode(&AsnPrefix); err != nil {
		return AsnResult{}, err
	}

	var total int
	for _, r := range AsnPrefix.Data.Prefixes {
		count, err := ComputeTotalIPs(r.Prefix)
		if err != nil {
			return AsnResult{}, err
		}
		total += count
	}

	result := AsnResult{
		ASN:      asn,
		Holder:   overview.Data.Holder,
		TotalIps: total,
	}

	for _, r := range AsnPrefix.Data.Prefixes {
		result.Prefixes = append(result.Prefixes, r.Prefix)
	}

	return result, nil
}

func Asn_Handler(w http.ResponseWriter, r *http.Request) {
	asn := chi.URLParam(r, "asn")
	data, err := FetchAsnData(asn)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
