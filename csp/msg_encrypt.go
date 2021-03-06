package csp

/*
#include "common.h"

extern CMSG_STREAM_INFO *mkStreamInfo(void *pvArg);

static CMSG_ENVELOPED_ENCODE_INFO *mkEnvelopedInfo(HCRYPTPROV hCryptProv, int cRecipients) {
    CRYPT_ALGORITHM_IDENTIFIER EncryptAlgorithm;
    memset(&EncryptAlgorithm, 0, sizeof(CRYPT_ALGORITHM_IDENTIFIER));

    EncryptAlgorithm.pszObjId = (LPSTR)ENCRYPT_OID;

	CMSG_ENVELOPED_ENCODE_INFO *res = malloc(sizeof(CMSG_ENVELOPED_ENCODE_INFO));
	memset(res, 0, sizeof(CMSG_ENVELOPED_ENCODE_INFO));

	res->cbSize = sizeof(CMSG_ENVELOPED_ENCODE_INFO);
	res->hCryptProv = hCryptProv;
	res->ContentEncryptionAlgorithm = EncryptAlgorithm;
	res->pvEncryptionAuxInfo = NULL;
	res->cRecipients = cRecipients;
	res->rgpRecipients = malloc(sizeof(PCERT_INFO) * cRecipients);
	memset(res->rgpRecipients, 0, sizeof(PCERT_INFO) * cRecipients);
	return res;
}

static void freeEnvelopedInfo(CMSG_ENVELOPED_ENCODE_INFO *info) {
	free(info->rgpRecipients);
	free(info);
}

static void setRecipientInfo(CMSG_ENVELOPED_ENCODE_INFO *out, int nSigner, PCCERT_CONTEXT pRecipientCert) {
	out->rgpRecipients[nSigner] = pRecipientCert->pCertInfo;
}

*/
import "C"
import (
	"fmt"
	"io"
	"unsafe"
)

// EncryptOptions specifies message encryption details
type EncryptOptions struct {
	Receivers []Cert // Receiving certificate list
}

// OpenToEncrypt creates new Msg in encrypt mode.
func OpenToEncrypt(dest io.Writer, options EncryptOptions) (*Msg, error) {
	if len(options.Receivers) == 0 {
		return nil, fmt.Errorf("Receivers certificates list is empty")
	}
	ctx, err := AcquireCtx("", "", ProvGost2012_512, CryptVerifyContext)
	if err != nil {
		return nil, err
	}
	res := new(Msg)
	res.callbackID = registerCallback(res.onWrite)
	si := C.mkStreamInfo(unsafe.Pointer(&res.callbackID))
	defer C.free(unsafe.Pointer(si))

	envelopedInfo := C.mkEnvelopedInfo(ctx.hProv, C.int(len(options.Receivers)))
	defer C.freeEnvelopedInfo(envelopedInfo)

	for i, receiverCert := range options.Receivers {
		C.setRecipientInfo(envelopedInfo, C.int(i), receiverCert.pCert)
	}
	res.w = dest
	res.hMsg = C.CryptMsgOpenToEncode(
		C.MY_ENC_TYPE,                 // encoding type
		0,                             // flags
		C.CMSG_ENVELOPED,              // message type
		unsafe.Pointer(envelopedInfo), // pointer to structure
		nil,                           // inner content OID
		si,                            // stream information
	)
	if res.hMsg == nil {
		return nil, getErr("Error opening message for encrypt")
	}
	return res, nil
}
