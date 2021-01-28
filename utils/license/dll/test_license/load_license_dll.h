#ifndef LOAD_LICENSE_DLL_H
#define LOAD_LICENSE_DLL_H
#include "../license.h"
#include <QLibrary>
#include <QDebug>

typedef int (*fGenerateFingerprint)(T_FINGERPRINT*);
typedef int (*fGenerateRsaKey)(T_PATH, T_PATH);
typedef int (*fSetPublicPemFile)(T_PATH);
typedef int (*fEncrypt)(T_FINGERPRINT, T_PATH, T_LICENSE*);
typedef int (*fSetPrivatePemFile)(T_PATH);
typedef int (*fDecrypt)(T_FINGERPRINT, T_PATH, T_LICENSE*);

static fGenerateFingerprint generateFingerPrint;
static fGenerateRsaKey generateRsaKey;
static fSetPublicPemFile setPublicPemFile;
static fEncrypt encrypt;
static fSetPrivatePemFile setPrivatePemFile;
static fDecrypt decrypt;

QLibrary* loadLicenseDll(const QString& dllPath) {
    QLibrary* library = new QLibrary(dllPath);
    if (!library->load()) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to load " << dllPath;
        library->deleteLater();
        return nullptr;
    }
    generateFingerPrint = (fGenerateFingerprint)library->resolve("GenerateFingerprint");
    generateRsaKey = (fGenerateRsaKey)library->resolve("GenerateRsaKey");
    setPublicPemFile = (fSetPublicPemFile)library->resolve("SetPublicPemFile");
    encrypt = (fEncrypt)library->resolve("Encrypt");
    setPrivatePemFile = (fSetPrivatePemFile)library->resolve("SetPrivatePemFile");
    decrypt = (fDecrypt)library->resolve("Decrypt");
    if (generateFingerPrint == nullptr
            || generateRsaKey == nullptr
            || setPublicPemFile == nullptr
            || encrypt == nullptr
            || setPrivatePemFile == nullptr
            || decrypt == nullptr) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to reslove function.";
        qDebug() << (void*)generateFingerPrint;
        qDebug() << (void*)generateRsaKey;
        qDebug() << (void*)setPublicPemFile;
        qDebug() << (void*)encrypt;
        qDebug() << (void*)setPrivatePemFile;
        qDebug() << (void*)decrypt;
        library->deleteLater();
        return nullptr;
    }
    return library;
}

#endif // LOAD_LICENSE_DLL_H
