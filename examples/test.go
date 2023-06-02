package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	"time"
)

func main() {
	req := url.NewRequest()
	req.Headers = url.ParseHeaders(`accept: application/json
x-apollo-cache-fetch-strategy: NETWORK_ONLY
x-apollo-expire-timeout: 0
x-apollo-expire-after-read: false
x-apollo-prefetch: false
x-apollo-cache-do-not-store: false
user-agent: Allegiant/6.9.7 Android/8.1.0 arm64-v8a
content-type: application/json; charset=utf-8
accept-encoding: gzip`)
	//req.Body = "{\"operationName\":\"getDepartureFlights\",\"variables\":{\"criteria\":{\"tripType\":\"ONEWAY\",\"origin\":\"LAS\",\"destination\":\"AVL\",\"departDate\":{\"date\":\"2023-10-15\",\"plusDays\":14,\"minusDays\":14},\"adultsCount\":1,\"lapInfantCount\":0,\"childrenCount\":0}},\"query\":\"query getDepartureFlights($criteria: FlightSearchCriteriaInput!) { transactionId flights(flightSearchCriteria: $criteria) { __typename departing { __typename ...FlightOption } } } fragment FlightOption on FlightOption { __typename id flight { __typename ...Flight } strikethruPrice price baseFare availableSeatsCount discountType totalDiscountValue incentiveType incentiveSavedAmount incentiveReturnDate onTimePerformance { __typename ...FlightOnTimePerformance } flightFeeForfeit { __typename ...FlightFeeForfeit } } fragment Flight on Flight { __typename id number departingTime arrivalTime isOvernight origin { __typename ...Airport } destination { __typename ...Airport } aircraft { __typename ...Aircraft } } fragment Airport on Airport { __typename code title displayName city country state geoPoint { __typename ...GeoPoint } } fragment GeoPoint on GeoPoint { __typename latitude longitude } fragment Aircraft on Aircraft { __typename make model } fragment FlightOnTimePerformance on FlightOnTimePerformance { __typename onTimeArrival thirtyMinuteLate cancellations disclaimer } fragment FlightFeeForfeit on FlightFeeForfeit { __typename flightForfiet ccvCharge { __typename ...MoneyTotal } } fragment MoneyTotal on MoneyTotal { __typename amount currency }\"}"
	//req.Body = "{\"operationName\":\"flights\",\"variables\":{\"flightSearchCriteria\":{\"tripType\":\"ONEWAY\",\"origin\":\"LAS\",\"destination\":\"AVL\",\"departDate\":{\"date\":\"2023-10-15\",\"minusDays\":14,\"plusDays\":14},\"adultsCount\":1,\"childrenCount\":0,\"lapInfantCount\":0,\"lapInfantDobs\":[]}},\"query\":\"query flights($flightSearchCriteria: FlightSearchCriteriaInput!) {\\n  transactionId\\n  flights(flightSearchCriteria: $flightSearchCriteria) {\\n    departing {\\n      ...FlightOptionFragment\\n      __typename\\n    }\\n    returning {\\n      ...FlightOptionFragment\\n      __typename\\n    }\\n    loyaltyFare {\\n      discount\\n      __typename\\n    }\\n    __typename\\n  }\\n  order {\\n    items {\\n      id\\n      __typename\\n      ... on FlightOrderItem {\\n        flight {\\n          id\\n          __typename\\n        }\\n        __typename\\n      }\\n      ... on ShowOrderItem {\\n        id\\n        type\\n        show {\\n          categoryCode\\n          __typename\\n        }\\n        __typename\\n      }\\n    }\\n    __typename\\n  }\\n}\\n\\nfragment FlightOptionFragment on FlightOption {\\n  id\\n  flight {\\n    id\\n    number\\n    origin {\\n      displayName\\n      code\\n      __typename\\n    }\\n    destination {\\n      code\\n      displayName\\n      __typename\\n    }\\n    departingTime\\n    arrivalTime\\n    isOvernight\\n    __typename\\n  }\\n  strikethruPrice\\n  price\\n  baseFare\\n  availableSeatsCount\\n  discountType\\n  totalDiscountValue\\n  incentiveType\\n  incentiveSavedAmount\\n  incentiveReturnDate\\n  onTimePerformance {\\n    onTimeArrival\\n    thirtyMinuteLate\\n    cancellations\\n    disclaimer\\n    __typename\\n  }\\n  __typename\\n}\\n\"}"
	req.Body = "{\"operationName\":\"flightMarket\",\"variables\":{\"origin\":\"LAS\",\"destination\":\"AVL\"},\"query\":\"query flightMarket($origin: IataCode!, $destination: IataCode!) {\\n  flightMarket(origin: $origin, destination: $destination) {\\n    type\\n    calendar {\\n      availableFrom\\n      availableUntil\\n      departingDates {\\n        date\\n        __typename\\n      }\\n      returningDates {\\n        date\\n        __typename\\n      }\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\"}"
	//es := &transport.Extensions{
	//	SupportedSignatureAlgorithms: []string{
	//		"ECDSAWithP256AndSHA256",
	//		"PKCS1WithSHA256",
	//		"ECDSAWithP384AndSHA384",
	//		"PKCS1WithSHA384",
	//		"ECDSAWithP521AndSHA512",
	//		"PKCS1WithSHA512",
	//		"PKCS1WithSHA1",
	//	},
	//	NotUsedGREASE: true,
	//}
	//es := &transport.Extensions{
	//	SupportedSignatureAlgorithms: []string{
	//		"ECDSAWithP256AndSHA256",
	//		"PSSWithSHA256",
	//		"PKCS1WithSHA256",
	//		"ECDSAWithP384AndSHA384",
	//		"PSSWithSHA384",
	//		"PKCS1WithSHA384",
	//		"PSSWithSHA512",
	//		"PKCS1WithSHA512",
	//		"PKCS1WithSHA1",
	//	},
	//	NotUsedGREASE: true,
	//}
	es := &transport.Extensions{
		SupportedSignatureAlgorithms: []string{
			"ECDSAWithP256AndSHA256",
			"PSSWithSHA256",
			"PKCS1WithSHA256",
			"ECDSAWithP384AndSHA384",
			"PSSWithSHA384",
			"PKCS1WithSHA384",
			"PSSWithSHA512",
			"PKCS1WithSHA512",
		},
		NotUsedGREASE: true,
	}
	tes := transport.ToTLSExtensions(es)
	req.TLSExtensions = tes
	//h2s := &transport.H2Settings{
	//	Settings: map[string]int{
	//		"HEADER_TABLE_SIZE":      4096,
	//		"ENABLE_PUSH":            1,
	//		"MAX_CONCURRENT_STREAMS": 0,
	//		"INITIAL_WINDOW_SIZE":    16777216,
	//		"MAX_FRAME_SIZE":         16384,
	//		"MAX_HEADER_LIST_SIZE":   0,
	//	},
	//	SettingsOrder: []string{
	//		"HEADER_TABLE_SIZE",
	//		"ENABLE_PUSH",
	//		"MAX_CONCURRENT_STREAMS",
	//		"INITIAL_WINDOW_SIZE",
	//		"MAX_FRAME_SIZE",
	//		"MAX_HEADER_LIST_SIZE",
	//	},
	//	ConnectionFlow: 16711680,
	//}
	//h2ss := transport.ToHTTP2Settings(h2s)
	//req.HTTP2Settings = h2ss
	// 第一组ciphers指纹
	//utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
	//utls.TLS_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_RSA_WITH_AES_256_CBC_SHA
	// 第二组ciphers指纹
	//utls.TLS_AES_128_GCM_SHA256
	//utls.TLS_AES_256_GCM_SHA384
	//utls.TLS_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
	//utls.TLS_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_RSA_WITH_AES_256_CBC_SHA
	// 第三组ciphers指纹
	//utls.TLS_AES_128_GCM_SHA256
	//utls.TLS_AES_256_GCM_SHA384
	//utls.TLS_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	//utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
	//utls.TLS_RSA_WITH_AES_128_GCM_SHA256
	//utls.TLS_RSA_WITH_AES_256_GCM_SHA384
	//utls.TLS_RSA_WITH_AES_128_CBC_SHA
	//utls.TLS_RSA_WITH_AES_256_CBC_SHA
	// 第一组ja3指纹
	//req.Ja3 = "771,49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0"
	// 第二组ja3指纹
	//23-65281-35-16-21
	req.Ja3 = "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43,29-23-24,0"
	// 第三组指纹
	//req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,17513-5-45-13-16-35-27-11-23-0-18-10-51-43-65281-21,29-23-24,0"
	//req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,10-23-51-13-16-5-11-17513-45-0-35-43-27-18-65281-21,29-23-24,0"
	//req.Proxies = "http://127.0.0.1:32768"
	req.Proxies = "http://metanfinite:Yyflight123_country-us_forcelocation-1@geo.iproyal.com:12321"
	req.Timeout = time.Duration(60) * time.Second
	//response, err := requests.Get("https://tls.peet.ws/api/all", req)
	//response, err := requests.Get("https://kawayiyi.com/tls", req)
	response, err := requests.Post("https://www.allegiantair.com/graphql", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.Text)
	fmt.Println(response.Headers)
	fmt.Println(response.StatusCode)
}
