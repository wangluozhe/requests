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
	//req.Body = "{\"operationName\":\"flights\",\"variables\":{\"flightSearchCriteria\":{\"tripType\":\"ONEWAY\",\"origin\":\"LAS\",\"destination\":\"AVL\",\"departDate\":{\"date\":\"2023-10-15\",\"minusDays\":14,\"plusDays\":14},\"adultsCount\":1,\"childrenCount\":0,\"lapInfantCount\":0,\"lapInfantDobs\":[]}},\"query\":\"query flights($flightSearchCriteria: FlightSearchCriteriaInput!) {\\n  transactionId\\n  flights(flightSearchCriteria: $flightSearchCriteria) {\\n    departing {\\n      ...FlightOptionFragment\\n      __typename\\n    }\\n    returning {\\n      ...FlightOptionFragment\\n      __typename\\n    }\\n    loyaltyFare {\\n      discount\\n      __typename\\n    }\\n    __typename\\n  }\\n  order {\\n    items {\\n      id\\n      __typename\\n      ... on FlightOrderItem {\\n        flight {\\n          id\\n          __typename\\n        }\\n        __typename\\n      }\\n      ... on ShowOrderItem {\\n        id\\n        type\\n        show {\\n          categoryCode\\n          __typename\\n        }\\n        __typename\\n      }\\n    }\\n    __typename\\n  }\\n}\\n\\nfragment FlightOptionFragment on FlightOption {\\n  id\\n  flight {\\n    id\\n    number\\n    origin {\\n      displayName\\n      code\\n      __typename\\n    }\\n    destination {\\n      code\\n      displayName\\n      __typename\\n    }\\n    departingTime\\n    arrivalTime\\n    isOvernight\\n    __typename\\n  }\\n  strikethruPrice\\n  price\\n  baseFare\\n  availableSeatsCount\\n  discountType\\n  totalDiscountValue\\n  incentiveType\\n  incentiveSavedAmount\\n  incentiveReturnDate\\n  onTimePerformance {\\n    onTimeArrival\\n    thirtyMinuteLate\\n    cancellations\\n    disclaimer\\n    __typename\\n  }\\n  __typename\\n}\\n\"}"
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
	}
	tes := transport.ToTLSExtensions(es)
	req.TLSExtensions = tes
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
	// 第三组指纹
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,10-23-51-13-16-5-11-17513-45-0-35-43-27-18-65281-21,29-23-24,0"
	//req.Proxies = "http://127.0.0.1:32768"
	req.Proxies = "http://metanfinite:Yyflight123_country-us_forcelocation-1@geo.iproyal.com:12321"
	req.Timeout = time.Duration(60) * time.Second
	response, err := requests.Get("https://tls.peet.ws/api/all", req)
	//response, err := requests.Post("https://www.allegiantair.com/graphql", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.Text)
	fmt.Println(response.Headers)
	fmt.Println(response.StatusCode)
}
