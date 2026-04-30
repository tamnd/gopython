package pycore

import (
	"fmt"
	"runtime"
	"sync"
)

const (
	PYReleaseLevelAlpha = 0xA
	PYReleaseLevelBeta  = 0xB
	PYReleaseLevelGamma = 0xC
	PYReleaseLevelFinal = 0xF

	PYMajorVersion  = 3
	PYMinorVersion  = 14
	PYMicroVersion  = 4
	PYReleaseLevel  = PYReleaseLevelFinal
	PYReleaseSerial = 0
	PYVersion       = "3.14.4+"
	PYVersionHex    = uint32(
		((PYMajorVersion & 0xff) << 24) |
			((PYMinorVersion & 0xff) << 16) |
			((PYMicroVersion & 0xff) << 8) |
			((PYReleaseLevel & 0xf) << 4) |
			((PYReleaseSerial & 0xf) << 0),
	)
)

const copyrightText = "" +
	"Copyright (c) 2001 Python Software Foundation.\n" +
	"All Rights Reserved.\n" +
	"\n" +
	"Copyright (c) 2000 BeOpen.com.\n" +
	"All Rights Reserved.\n" +
	"\n" +
	"Copyright (c) 1995-2001 Corporation for National Research Initiatives.\n" +
	"All Rights Reserved.\n" +
	"\n" +
	"Copyright (c) 1991-1995 Stichting Mathematisch Centrum, Amsterdam.\n" +
	"All Rights Reserved."

var (
	versionOnce sync.Once
	versionText string

	buildInfoOnce sync.Once
	buildInfoText string
)

func PyGetCopyright() string {
	return copyrightText
}

func PyGetCompiler() string {
	return "[Go " + runtime.Version() + "]"
}

func PyGetPlatform() string {
	return runtime.GOOS
}

func PyGetBuildInfo() string {
	buildInfoOnce.Do(func() {
		buildInfoText = "main, Jan 01 1970, 00:00:00"
	})
	return buildInfoText
}

func PyGetVersion() string {
	versionOnce.Do(func() {
		versionText = fmt.Sprintf(
			"%.80s (%.80s) %.80s",
			PYVersion,
			PyGetBuildInfo(),
			PyGetCompiler(),
		)
	})
	return versionText
}

func PyFPEdummy(_ any) float64 {
	return 1.0
}
