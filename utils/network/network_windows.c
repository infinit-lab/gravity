#ifdef WINDOWS

#include "network.h"
#include <stdio.h>
#include <Windows.h>
#include <iphlpapi.h>
#include <string.h>

PIP_ADAPTER_INFO GetAdapters();

PIP_ADAPTER_INFO GetAdapters() {
    PIP_ADAPTER_INFO pIpAdapterInfo;
    DWORD nSize;

    nSize = sizeof(IP_ADAPTER_INFO);
    pIpAdapterInfo = malloc(nSize);

    int nRet = GetAdaptersInfo(pIpAdapterInfo, &nSize);
    if (ERROR_BUFFER_OVERFLOW == nRet) {
        free(pIpAdapterInfo);
        pIpAdapterInfo = malloc(nSize);
        nRet = GetAdaptersInfo(pIpAdapterInfo, &nSize);
    }
    if (ERROR_SUCCESS != nRet) {
        free(pIpAdapterInfo);
        return 0;
    }
    return pIpAdapterInfo;
}

int GetAdaptersCount() {
    PIP_ADAPTER_INFO pIpAdapterInfo, pAdapter;
    int nAdaptersCount;

    pIpAdapterInfo = GetAdapters();
    if (pIpAdapterInfo == 0) {
        return 0;
    }
    pAdapter = pIpAdapterInfo;
    nAdaptersCount = 0;

    while (pAdapter) {
        ++nAdaptersCount;
        pAdapter = pAdapter->Next;
    }
    free(pIpAdapterInfo);
    return nAdaptersCount;
}

int GetAdapter(int i, PT_ADAPTER adapter) {
    PIP_ADAPTER_INFO pIpAdapterInfo, pAdapter;
    int nAdaptersCount;

    if (adapter == 0) {
        return NET_ERROR_FAILED;
    }

    pIpAdapterInfo = GetAdapters();
    if (pIpAdapterInfo == 0) {
        return NET_ERROR_FAILED;
    }
    pAdapter = pIpAdapterInfo;
    nAdaptersCount = 0;
    while (pAdapter) {
        if (nAdaptersCount == i) {
            memset(adapter, 0, sizeof(T_ADAPTER));
            adapter->index = pAdapter->Index;
            strncpy(adapter->name, pAdapter->AdapterName, NAME_LENGTH - 1);
            strncpy(adapter->description, pAdapter->Description, NAME_LENGTH - 1);
            for (int i = 0; i < pAdapter->AddressLength; ++i) {
                char temp[4];
                memset(temp, 0, sizeof(temp));
                if (i < pAdapter->AddressLength - 1) {
                    sprintf(temp, "%02x-", pAdapter->Address[i]);
                    strcat(adapter->mac, temp);
                } else {
                    sprintf(temp, "%02x", pAdapter->Address[i]);
                    strcat(adapter->mac, temp);
                }
            }
            PIP_ADDR_STRING pIpAddr = &pAdapter->IpAddressList;
            if (pIpAddr) {
                strncpy(adapter->ip, pIpAddr->IpAddress.String, ADDR_LENGTH - 1);
                strncpy(adapter->mask, pIpAddr->IpMask.String, ADDR_LENGTH - 1);
            }
            PIP_ADDR_STRING pGateway = &pAdapter->GatewayList;
            if (pGateway) {
                strncpy(adapter->gateway, pGateway->IpAddress.String, ADDR_LENGTH - 1);
            }
            break;
        }
        ++nAdaptersCount;
        pAdapter = pAdapter->Next;
    }
    free(pIpAdapterInfo);
    return NET_ERROR_SUCCESS;
}

#endif
