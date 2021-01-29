#include <QCoreApplication>
#include <QDateTime>
#include "load_license_dll.h"

int main(int argc, char *argv[])
{
    int ret = 0;
    T_PATH publicPem = {0}, privatePem = {0}, licensePath = {0};
    T_FINGERPRINT fingerprint = {0};
    T_LICENSE *license = nullptr, *decryptLicense = nullptr;

    license = new(T_LICENSE);
    memset(license, 0, sizeof(T_LICENSE));
    decryptLicense = new(T_LICENSE);
    memset(decryptLicense, 0, sizeof(T_LICENSE));

    QCoreApplication a(argc, argv);
    QLibrary* library = loadLicenseDll("../../license.dll");
    if (!library) {
        goto failed;
    }

    sprintf(publicPem.path, "public.pem");
    sprintf(privatePem.path, "private.pem");
    // 准备：生成公钥与私钥，公钥用于加密，私钥用于解密, 客户机需保存公钥，注册机需保存私钥
    ret = generateRsaKey(publicPem, privatePem);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to generateRsaKey";
        goto failed;
    }

    // 第一步：设置公钥用于加密机器指纹
    ret = setPublicPemFile(publicPem);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to setPublicPemFile";
        goto failed;
    }

    // 第二步：获取客户机的机器指纹
    ret = generateFingerPrint(&fingerprint);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to generateFingerprint";
        goto failed;
    }
    qDebug() << __FUNCTION__ << __LINE__ << fingerprint.fingerprint;

    license->isForever = 0;
    sprintf(license->validDatetime, "2021-01-28 10:00:00");
    sprintf(license->currentDatetime, "2021-01-28 10:00:00");
    license->validDuration = 115200;
    license->currentDuration = 9600;
    for (int i = 0; i < 2; i++) {
        sprintf(license->authorizations[i].authType, "authType");
        sprintf(license->authorizations[i].name, "name");
        sprintf(license->authorizations[i].valueType, "valueType");
        for (int j = 0; j < 2; j++) {
            sprintf(license->authorizations[i].values[j].value, "value");
        }
        sprintf(license->authorizations[i].current, "current");
    }

    // 第三步：设置私钥用于解密机器指纹
    ret = setPrivatePemFile(privatePem);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to setPrivatePemFile";
        goto failed;
    }

    // 第四步：加密生成证书文件
    sprintf(licensePath.path, "license.lic");
    ret = encrypt(fingerprint, licensePath, license);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to encrypt";
        goto failed;
    }

    // 第五步：解密证书文件获取证书内容
    ret = decrypt(licensePath, decryptLicense);
    if (ret == ERROR_FAIL) {
        qDebug() << __FUNCTION__ << __LINE__ << "Failed to decrypt";
        goto failed;
    }

    qDebug() << decryptLicense->isForever;
    qDebug() << decryptLicense->validDatetime;
    qDebug() << decryptLicense->currentDatetime;
    qDebug() << decryptLicense->validDuration;
    qDebug() << decryptLicense->currentDuration;
    for (int i = 0; i < AUTHORIZATION_LENGTH; i++) {
        if (decryptLicense->authorizations[i].authType[0] == 0) {
            break;
        }
        qDebug() << decryptLicense->authorizations[i].authType;
        qDebug() << decryptLicense->authorizations[i].name;
        qDebug() << decryptLicense->authorizations[i].valueType;
        for (int j = 0; j < VALUE_LENGTH; j++) {
            if (decryptLicense->authorizations[i].values[j].value[0] == 0) {
                break;
            }
            qDebug() << decryptLicense->authorizations[i].values[j].value;
        }
        qDebug() << decryptLicense->authorizations[i].current;
    }

    return a.exec();
failed:
    if (license) {
        delete license;
    }
    if (decryptLicense) {
        delete decryptLicense;
    }
    system("pause");
    return -1;
}
