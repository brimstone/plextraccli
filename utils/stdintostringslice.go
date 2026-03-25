// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package utils

import (
	"bufio"
	"os"
)

func StdinToStringSlice() ([]string, error) {
	var ret []string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return ret, nil
}
