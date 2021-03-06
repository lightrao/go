// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpu implements processor feature detection
// used by the Go standard library.
package cpu

// DebugOptions is set to true by the runtime if the OS supports GODEBUGCPU.
// This should not be changed after it is initialized.
var DebugOptions bool

// CacheLinePad is used to pad structs to avoid false sharing.
type CacheLinePad struct{ _ [CacheLinePadSize]byte }

// CacheLineSize is the CPU's assumed cache line size.
// There is currently no runtime detection of the real cache line size
// so we use the constant per GOARCH CacheLinePadSize as an approximation.
var CacheLineSize uintptr = CacheLinePadSize

var X86 x86

// The booleans in x86 contain the correspondingly named cpuid feature bit.
// HasAVX and HasAVX2 are only set if the OS does support XMM and YMM registers
// in addition to the cpuid feature bit being set.
// The struct is padded to avoid false sharing.
type x86 struct {
	_            CacheLinePad
	HasAES       bool
	HasADX       bool
	HasAVX       bool
	HasAVX2      bool
	HasBMI1      bool
	HasBMI2      bool
	HasERMS      bool
	HasFMA       bool
	HasOSXSAVE   bool
	HasPCLMULQDQ bool
	HasPOPCNT    bool
	HasSSE2      bool
	HasSSE3      bool
	HasSSSE3     bool
	HasSSE41     bool
	HasSSE42     bool
	_            CacheLinePad
}

var PPC64 ppc64

// For ppc64x, it is safe to check only for ISA level starting on ISA v3.00,
// since there are no optional categories. There are some exceptions that also
// require kernel support to work (darn, scv), so there are feature bits for
// those as well. The minimum processor requirement is POWER8 (ISA 2.07), so we
// maintain some of the old feature checks for optional categories for
// safety.
// The struct is padded to avoid false sharing.
type ppc64 struct {
	_          CacheLinePad
	HasVMX     bool // Vector unit (Altivec)
	HasDFP     bool // Decimal Floating Point unit
	HasVSX     bool // Vector-scalar unit
	HasHTM     bool // Hardware Transactional Memory
	HasISEL    bool // Integer select
	HasVCRYPTO bool // Vector cryptography
	HasHTMNOSC bool // HTM: kernel-aborted transaction in syscalls
	HasDARN    bool // Hardware random number generator (requires kernel enablement)
	HasSCV     bool // Syscall vectored (requires kernel enablement)
	IsPOWER8   bool // ISA v2.07 (POWER8)
	IsPOWER9   bool // ISA v3.00 (POWER9)
	_          CacheLinePad
}

var ARM arm

// The booleans in arm contain the correspondingly named cpu feature bit.
// The struct is padded to avoid false sharing.
type arm struct {
	_        CacheLinePad
	HasVFPv4 bool
	HasIDIVA bool
	_        CacheLinePad
}

var ARM64 arm64

// The booleans in arm64 contain the correspondingly named cpu feature bit.
// The struct is padded to avoid false sharing.
type arm64 struct {
	_           CacheLinePad
	HasFP       bool
	HasASIMD    bool
	HasEVTSTRM  bool
	HasAES      bool
	HasPMULL    bool
	HasSHA1     bool
	HasSHA2     bool
	HasCRC32    bool
	HasATOMICS  bool
	HasFPHP     bool
	HasASIMDHP  bool
	HasCPUID    bool
	HasASIMDRDM bool
	HasJSCVT    bool
	HasFCMA     bool
	HasLRCPC    bool
	HasDCPOP    bool
	HasSHA3     bool
	HasSM3      bool
	HasSM4      bool
	HasASIMDDP  bool
	HasSHA512   bool
	HasSVE      bool
	HasASIMDFHM bool
	_           CacheLinePad
}

var S390X s390x

type s390x struct {
	_               CacheLinePad
	HasZArch        bool // z architecture mode is active [mandatory]
	HasSTFLE        bool // store facility list extended [mandatory]
	HasLDisp        bool // long (20-bit) displacements [mandatory]
	HasEImm         bool // 32-bit immediates [mandatory]
	HasDFP          bool // decimal floating point
	HasETF3Enhanced bool // ETF-3 enhanced
	HasMSA          bool // message security assist (CPACF)
	HasAES          bool // KM-AES{128,192,256} functions
	HasAESCBC       bool // KMC-AES{128,192,256} functions
	HasAESCTR       bool // KMCTR-AES{128,192,256} functions
	HasAESGCM       bool // KMA-GCM-AES{128,192,256} functions
	HasGHASH        bool // KIMD-GHASH function
	HasSHA1         bool // K{I,L}MD-SHA-1 functions
	HasSHA256       bool // K{I,L}MD-SHA-256 functions
	HasSHA512       bool // K{I,L}MD-SHA-512 functions
	HasVX           bool // vector facility. Note: the runtime sets this when it processes auxv records.
	_               CacheLinePad
}

// Initialize examines the processor and sets the relevant variables above.
// This is called by the runtime package early in program initialization,
// before normal init functions are run. env is set by runtime if the OS supports GODEBUGCPU.
func Initialize(env string) {
	doinit()
	processOptions(env)
}

// options contains the cpu debug options that can be used in GODEBUGCPU.
// Options are arch dependent and are added by the arch specific doinit functions.
// Features that are mandatory for the specific GOARCH should not be added to options
// (e.g. SSE2 on amd64).
var options []option

// Option names should be lower case. e.g. avx instead of AVX.
type option struct {
	Name      string
	Feature   *bool
	Specified bool // Stores if feature value was specified in GODEBUGCPU.
	Enable    bool // Stores if feature should be enabled.
}

// processOptions enables or disables CPU feature values based on the parsed env string.
// The env string is expected to be of the form feature1=value1,feature2=value2...
// where feature names is one of the architecture specifc list stored in the
// cpu packages options variable and values are either 'on' or 'off'.
// If env contains all=off then all cpu features referenced through the options
// variable are disabled. Other feature names and values result in warning messages.
func processOptions(env string) {
field:
	for env != "" {
		field := ""
		i := indexByte(env, ',')
		if i < 0 {
			field, env = env, ""
		} else {
			field, env = env[:i], env[i+1:]
		}
		i = indexByte(field, '=')
		if i < 0 {
			print("GODEBUGCPU: no value specified for \"", field, "\"\n")
			continue
		}
		key, value := field[:i], field[i+1:]

		var enable bool
		switch value {
		case "on":
			enable = true
		case "off":
			enable = false
		default:
			print("GODEBUGCPU: value \"", value, "\" not supported for option ", key, "\n")
			continue field
		}

		if key == "all" {
			for i := range options {
				options[i].Specified = true
				options[i].Enable = enable
			}
			continue field
		}

		for i := range options {
			if options[i].Name == key {
				options[i].Specified = true
				options[i].Enable = enable
				continue field
			}
		}

		print("GODEBUGCPU: unknown cpu feature \"", key, "\"\n")
	}

	for _, o := range options {
		if !o.Specified {
			continue
		}

		if o.Enable && !*o.Feature {
			print("GODEBUGCPU: can not enable \"", o.Name, "\", missing hardware support\n")
			continue
		}

		*o.Feature = o.Enable
	}
}

// indexByte returns the index of the first instance of c in s,
// or -1 if c is not present in s.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
