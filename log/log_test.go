/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
 */
package log

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func logPrint() {
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")
	Fatal("fatal")
	Trace("trace")

	testValue := 1
	Debugf("debug %v", testValue)
	Infof("info %v", testValue)
	Warnf("warn %v", testValue)
	Errorf("error %v", testValue)
	Fatalf("fatal %v", testValue)
	Tracef("trace %v", testValue)
}

func TestLog(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	Init(PATH, Stdout)
	Log.SetDebugLevel(DebugLog)
	logPrint()

	Log.SetDebugLevel(WarnLog)

	logPrint()

	err := ClosePrintLog()
	assert.Nil(t, err)
}

func TestNewLogFile(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()
	Init(PATH, Stdout)
	logfileNum1, err1 := ioutil.ReadDir("Log/")
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	logPrint()
	isNeedNewFile := CheckIfNeedNewFile()
	assert.NotEqual(t, isNeedNewFile, true)
	ClosePrintLog()
	time.Sleep(time.Second * 2)
	Init(PATH, Stdout)
	logfileNum2, err2 := ioutil.ReadDir("Log/")
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	assert.Equal(t, len(logfileNum1), (len(logfileNum2) - 1))
}

func TestCheckIfNeedNewFile(t *testing.T) {
	raw, _ := hex.DecodeString("c3a14ebb3e35ad8d04fbd159559d85d5951ce03cc965ce0352610938d5fae3c6")

	aa, _ := common.Uint256ParseFromBytes(raw)
	fmt.Println(aa.ToHexString())
}
