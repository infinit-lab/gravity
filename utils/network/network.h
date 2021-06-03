#ifndef __NETWORK_H__
#define __NETWORK_H__

#define NAME_LENGTH (256)
#define ADDR_LENGTH (32)

#define NET_ERROR_SUCCESS 0L
#define NET_ERROR_FAILED 1L

int GetAdaptersCount();

typedef struct _T_ADAPTER {
  int index;
  char name[NAME_LENGTH];
  char description[NAME_LENGTH];
  char mac[ADDR_LENGTH];
  char ip[ADDR_LENGTH];
  char mask[ADDR_LENGTH];
  char gateway[ADDR_LENGTH];
}T_ADAPTER, *PT_ADAPTER;

int GetAdapter(int i, PT_ADAPTER pAdapter);

int SetAdapter(PT_ADAPTER pAdapter);

#endif //__NET_H__
