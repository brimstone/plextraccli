// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

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

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}
