// runit manages services usually found in /etc/service or /service, which
// are soft links to the actual service directories in /etc/sv:
//    <service_name>/
//        run
//        log/
//           run
//        supervise/
//           pid  # => 4994
//           stat # => run / down
//
package services

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

type Runit struct {
	path string
}

func DetectRunit(root string) (*Runit, error) {
	path := root + "etc/service"
	result, err := fileExists(path)
	if err != nil {
		return nil, err
	}

	if !result {
		path = root + "service"
		result, err = fileExists(path)
		if err != nil {
			return nil, err
		}
	}

	if !result {
		log.Println("runit not detected in /etc/service or /service")
		return nil, nil
	}

	matches, err := filepath.Glob(path + "/*/run")
	if err != nil {
		return nil, err
	}

	if len(matches) > 0 {
		log.Println("Detected runit in " + path)
		return &Runit{path}, nil
	}

	return nil, nil
}

func (r *Runit) FindServicePID(serviceName string) (int32, error) {
	matches, err := filepath.Glob(r.path + "/" + serviceName + "/run")
	if err != nil {
		return 0, err
	}

	if len(matches) == 0 {
		return -1, nil
	}

	content, err := ioutil.ReadFile(r.path + "/" + serviceName + "/supervise/pid")
	if len(content) == 0 {
		// service exists but is not running
		return 0, nil
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(pid), nil
}
