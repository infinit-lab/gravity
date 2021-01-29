package main

/*
#include <string.h>

#define STRING_MAX_LENGTH    (256)
#define VALUE_LENGTH         (256)
#define AUTHORIZATION_LENGTH (256)

#define ERROR_OK   (0)
#define ERROR_FAIL (1)

struct T_VALUE {
	char value[STRING_MAX_LENGTH];
};

struct T_AUTHORIZATION {
	char authType[STRING_MAX_LENGTH];
	char name[STRING_MAX_LENGTH];
	char valueType[STRING_MAX_LENGTH];
	struct T_VALUE values[STRING_MAX_LENGTH];
	char current[STRING_MAX_LENGTH];
};

struct T_LICENSE {
	int isForever;
	char validDatetime[STRING_MAX_LENGTH];
	char currentDatetime[STRING_MAX_LENGTH];
	int validDuration;
	int currentDuration;
	struct T_AUTHORIZATION authorizations[AUTHORIZATION_LENGTH];
};

struct T_FINGERPRINT {
	char fingerprint[STRING_MAX_LENGTH];
};

struct T_PATH {
	char path[STRING_MAX_LENGTH];
};
*/
import "C"
import (
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils"
	"github.com/infinit-lab/gravity/utils/license"
	"io/ioutil"
	"os"
	"unsafe"
)

//export GenerateFingerprint
func GenerateFingerprint(fingerprint *C.struct_T_FINGERPRINT) C.int {
	if fingerprint == nil {
		return C.ERROR_FAIL
	}
	f, err := license.GenerateFingerprint()
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	utils.GoStringToCArray(f, unsafe.Pointer(&fingerprint.fingerprint[0]), C.STRING_MAX_LENGTH)
	return C.ERROR_OK
}

//export GenerateRsaKey
func GenerateRsaKey(publicPemPath C.struct_T_PATH, privatePemPath C.struct_T_PATH) C.int {
	publicFile := utils.CArrayToGoString(unsafe.Pointer(&publicPemPath.path[0]), C.STRING_MAX_LENGTH)
	privateFile := utils.CArrayToGoString(unsafe.Pointer(&privatePemPath.path[0]), C.STRING_MAX_LENGTH)
	err := license.GenerateRsaKey(publicFile, privateFile)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	return C.ERROR_OK
}

//export SetPublicPemFile
func SetPublicPemFile(publicPemPath C.struct_T_PATH) C.int {
	publicFile := utils.CArrayToGoString(unsafe.Pointer(&publicPemPath.path[0]), C.STRING_MAX_LENGTH)
	err := license.SetPublicPem(publicFile)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	return C.ERROR_OK
}

//export Encrypt
func Encrypt(fingerprint C.struct_T_FINGERPRINT, path C.struct_T_PATH, lic *C.struct_T_LICENSE) C.int {
	f := utils.CArrayToGoString(unsafe.Pointer(&fingerprint.fingerprint[0]), C.STRING_MAX_LENGTH)
	p := utils.CArrayToGoString(unsafe.Pointer(&path.path[0]), C.STRING_MAX_LENGTH)
	var l license.License
	if lic.isForever == C.int(0) {
		l.IsForever = true
	} else {
		l.IsForever = false
	}
	l.ValidDatetime = utils.CArrayToGoString(unsafe.Pointer(&lic.validDatetime[0]), C.STRING_MAX_LENGTH)
	l.CurrentDatetime = utils.CArrayToGoString(unsafe.Pointer(&lic.currentDatetime[0]), C.STRING_MAX_LENGTH)
	l.ValidDuration = int(lic.validDuration)
	l.CurrentDuration = int(lic.currentDuration)
	for i := C.int(0); i < C.AUTHORIZATION_LENGTH; i++ {
		auth := lic.authorizations[i]
		if auth.authType[0] == C.char('0') {
			break
		}
		var a license.Authorization
		a.Type = utils.CArrayToGoString(unsafe.Pointer(&auth.authType[0]), C.STRING_MAX_LENGTH)
		a.Name = utils.CArrayToGoString(unsafe.Pointer(&auth.name[0]), C.STRING_MAX_LENGTH)
		a.ValueType = utils.CArrayToGoString(unsafe.Pointer(&auth.valueType[0]), C.STRING_MAX_LENGTH)
		for j := C.int(0); j < C.VALUE_LENGTH; j++ {
			value := auth.values[j]
			if value.value[0] == C.char('0') {
				break
			}
			v := utils.CArrayToGoString(unsafe.Pointer(&value.value[0]), C.STRING_MAX_LENGTH)
			a.Values = append(a.Values, v)
		}
		a.Current = utils.CArrayToGoString(unsafe.Pointer(&auth.current[0]), C.STRING_MAX_LENGTH)
		l.Authorizations = append(l.Authorizations, a)
	}
	content, err := license.GenerateLicense(f, l)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	err = ioutil.WriteFile(p, []byte(content), os.ModePerm)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	return C.ERROR_OK
}

//export SetPrivatePemFile
func SetPrivatePemFile(privatePemPath C.struct_T_PATH) C.int {
	privateFile := utils.CArrayToGoString(unsafe.Pointer(&privatePemPath.path[0]), C.STRING_MAX_LENGTH)
	err := license.SetPrivatePem(privateFile)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	return C.ERROR_OK
}

//export Decrypt
func Decrypt(path C.struct_T_PATH, lic *C.struct_T_LICENSE) C.int {
	if lic == nil {
		return C.ERROR_FAIL
	}
	p := utils.CArrayToGoString(unsafe.Pointer(&path.path[0]), C.STRING_MAX_LENGTH)
	content, err := ioutil.ReadFile(p)
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	l, err := license.LoadLicense(string(content))
	if err != nil {
		printer.Error(err)
		return C.ERROR_FAIL
	}
	C.memset(unsafe.Pointer(lic), 0, C.ulonglong(unsafe.Sizeof(*lic)))
	if l.IsForever {
		lic.isForever = C.int(1)
	} else {
		lic.isForever = C.int(0)
	}
	utils.GoStringToCArray(l.ValidDatetime, unsafe.Pointer(&lic.validDatetime[0]), C.STRING_MAX_LENGTH)
	utils.GoStringToCArray(l.CurrentDatetime, unsafe.Pointer(&lic.currentDatetime[0]), C.STRING_MAX_LENGTH)
	lic.validDuration = C.int(l.ValidDuration)
	lic.currentDuration = C.int(l.CurrentDuration)
	for i := 0; i < len(l.Authorizations) && i < int(C.AUTHORIZATION_LENGTH); i++ {
		auth := &lic.authorizations[C.int(i)]
		utils.GoStringToCArray(l.Authorizations[i].Type, unsafe.Pointer(&auth.authType[0]), C.STRING_MAX_LENGTH)
		utils.GoStringToCArray(l.Authorizations[i].Name, unsafe.Pointer(&auth.name[0]), C.STRING_MAX_LENGTH)
		utils.GoStringToCArray(l.Authorizations[i].ValueType, unsafe.Pointer(&auth.valueType[0]), C.STRING_MAX_LENGTH)
		for j := 0; j < len(l.Authorizations[i].Values) && j < int(C.VALUE_LENGTH); j++ {
			value := &auth.values[C.int(j)]
			utils.GoStringToCArray(l.Authorizations[i].Values[j], unsafe.Pointer(&value.value[0]), C.STRING_MAX_LENGTH)
		}
		utils.GoStringToCArray(l.Authorizations[i].Current, unsafe.Pointer(&auth.current[0]), C.STRING_MAX_LENGTH)
	}
	return C.ERROR_OK
}

func main() {

}
