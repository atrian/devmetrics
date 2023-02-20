package middlewares

import (
	"log"
	"net"
	"net/http"
)

func TrustedSubnetMW(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// При пустом значении переменной trusted_subnet метрики должны обрабатываться сервером без дополнительных ограничений
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Если парсинг безопасной сети выдает ошибку, передаем запрос дальше
			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Если адрес входит в список доверенных, передаем запрос дальше
			agentIp := net.ParseIP(r.Header.Get("X-Real-IP"))
			log.Printf("agentIp: %v, trusted network: %v", agentIp.String(), ipNet.String())
			if agentIp != nil && ipNet.Contains(agentIp) {
				next.ServeHTTP(w, r)
				return
			}

			// иначе сбрасываем соединение
			log.Printf("Drop connection for Ip: %v", agentIp.String())
			dropConnection(w)
		})
	}
}

func dropConnection(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}
