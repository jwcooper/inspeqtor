package daemon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/mperham/inspeqtor/metrics"
	"github.com/mperham/inspeqtor/util"
)

/*
 To test:
 brew tap homebrew/nginx
 brew install nginx-full --with-status
 Now place this in /usr/local/etc/nginx/nginx.conf within the server{} block:
   location /status {
     stub_status on;
     access_log off;
     allow 127.0.0.1;
     deny all;
   }
*/

func init() {
	metrics.Sources["nginx"] = buildNginxSource
}

type nginxSource struct {
	Hostname string
	Port     string
	Endpoint string
	metrics  map[string]bool
	args     []string
	client   func(string, string, string) ([]byte, error)
}

func (rs *nginxSource) Name() string {
	return "nginx"
}

func (rs *nginxSource) Watch(metricName string) {
	rs.metrics[metricName] = true
}

func (rs *nginxSource) Capture() (metrics.Map, error) {
	return rs.runCli()
}

func (rs *nginxSource) Prepare() error {
	return nil
}

func (rs *nginxSource) ValidMetrics() []metrics.Descriptor {
	return nginxMetrics
}

func defaultClient(host string, port string, ep string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%s%s", host, port, ep)
	util.Debug("Fetching nginx status from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

var (
	digits = regexp.MustCompile(`(\d+)`)
)

func (rs *nginxSource) runCli() (metrics.Map, error) {
	sout, err := rs.client(rs.Hostname, rs.Port, rs.Endpoint)
	if err != nil {
		return nil, err
	}
	if sout[0] != 0x41 { // first char should be 'A'
		util.Warn(string(sout))
		return nil, errors.New("Unknown nginx status output")
	}

	values := map[string]float64{}
	results := digits.FindAllStringSubmatch(string(sout), 7)
	if results == nil || len(results) != 7 {
		return nil, errors.New("Unknown nginx input")
	}

	for idx, met := range nginxMetrics {
		if !rs.metrics[met.Name] {
			continue
		}
		val, err := strconv.ParseInt(results[idx][0], 10, 64)
		if err != nil {
			return nil, err
		}
		values[met.Name] = float64(val)
	}

	if len(rs.metrics) > len(values) {
		for k := range rs.metrics {
			if _, ok := values[k]; !ok {
				util.Info("Could not find metric %s(%s), did you spell it right?", rs.Name(), k)
			}
		}
	}

	return values, nil
}

func buildNginxSource(params map[string]string) (metrics.Source, error) {
	rs := &nginxSource{"localhost", "80", "/status", map[string]bool{}, nil, defaultClient}
	for k, v := range params {
		switch k {
		case "endpoint":
			rs.Endpoint = v
		case "hostname":
			rs.Hostname = v
		case "port":
			_, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, err
			}
			rs.Port = v
		}
	}
	return rs, nil
}

var (
	nginxMetrics = []metrics.Descriptor{
		metrics.D("Active_connections", g, nil, nil),
		metrics.D("accepts", c, nil, nil),
		metrics.D("handled", c, nil, nil),
		metrics.D("requests", c, nil, nil),
		metrics.D("Reading", g, nil, nil),
		metrics.D("Writing", g, nil, nil),
		metrics.D("Waiting", g, nil, nil),
	}
)
