#ifdef LINUX

#include "network.h"
#include <stdio.h>
#include <string.h>
#include <net/if.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <errno.h>
#include <unistd.h>
#include <net/route.h>
#include <netinet/in.h>

int GetAdaptersCount() {
	int fd = 0;
	int interfaceNum = 0;
	struct ifreq buf[16] = {0};
	struct ifconf ifc = {0};

	if ((fd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
		close(fd);
		return 0;
	}

	ifc.ifc_len = sizeof(buf);
	ifc.ifc_buf = (caddr_t)buf;
	if (!ioctl(fd, SIOCGIFCONF, (char *)&ifc)) {
		interfaceNum = ifc.ifc_len / sizeof(struct ifreq);
	}
	close(fd);
	return interfaceNum;
}

int GetAdapter(int i, PT_ADAPTER pAdapter) {
	if (!pAdapter) {
		return NET_ERROR_SUCCESS;
	}
	int fd = 0;
	int interfaceNum = 0;
	struct ifreq buf[16] = {0};
	struct ifconf ifc = {0};
	char ip[32] = {0};
	char broadAddr[32] = {0};
	char subnetMask[32] = {0};

	if ((fd = socket(AF_INET, SOCK_DGRAM, 0)) < 0) {
		return NET_ERROR_FAILED;
	}

	ifc.ifc_len = sizeof(buf);
	ifc.ifc_buf = (caddr_t)buf;
	if (ioctl(fd, SIOCGIFCONF, (char *)&ifc) != 0) {
		goto failed;
	}
	interfaceNum = ifc.ifc_len / sizeof(struct ifreq);
	if (i >= interfaceNum) {
		goto failed;
	}
	strncpy(pAdapter->name, buf[i].ifr_name, NAME_LENGTH - 1);

	if (ioctl(fd, SIOCGIFHWADDR, (char*)(&buf[i])) != 0) {
		goto failed;
	}
	memset(pAdapter->mac, 0, sizeof(pAdapter->mac));
	sprintf(pAdapter->mac, "%02x-%02x-%02x-%02x-%02x-%02x",
		(unsigned char)buf[i].ifr_hwaddr.sa_data[0],
		(unsigned char)buf[i].ifr_hwaddr.sa_data[1],
		(unsigned char)buf[i].ifr_hwaddr.sa_data[2],
		(unsigned char)buf[i].ifr_hwaddr.sa_data[3],
		(unsigned char)buf[i].ifr_hwaddr.sa_data[4],
		(unsigned char)buf[i].ifr_hwaddr.sa_data[5]);

	if (ioctl(fd, SIOCGIFADDR, (char*)(&buf[i])) != 0) {
		goto failed;
	}
	sprintf(pAdapter->ip, "%s", (char*)inet_ntoa(((struct sockaddr_in*)&(buf[i].ifr_addr))->sin_addr));

	if (ioctl(fd, SIOCGIFNETMASK, (char*)(&buf[i])) != 0) {
		goto failed;
	}
	sprintf(pAdapter->mask, "%s", (char*)inet_ntoa(((struct sockaddr_in*)&(buf[i].ifr_netmask))->sin_addr));

	close(fd);
	return NET_ERROR_SUCCESS;
failed:
	close(fd);
	return NET_ERROR_FAILED;
}

int SetAdapter(PT_ADAPTER pAdapter) {
	int fd = 0;
	int ret = 0;
	struct ifreq ifr = {0};
	struct sockaddr_in *sin = {0};
	struct rtentry rt = {0};

	fd = socket(AF_INET, SOCK_DGRAM, 0);
	if (fd < 0) {
		return NET_ERROR_FAILED;
	}
	printf("SetAdapter Name is %s\n", pAdapter->name);
	strcpy(ifr.ifr_name, pAdapter->name);
	sin = (struct sockaddr_in*)&ifr.ifr_addr;
	sin->sin_family = AF_INET;

	printf("SetAdapter Ip is %s\n", pAdapter->ip);
	if (ret = inet_aton(pAdapter->ip, &(sin->sin_addr)) == 0) {
		printf("SetAdapter inet_aton error is %d\n", ret);
		goto failed;
	}
	if (ret = ioctl(fd, SIOCSIFADDR, &ifr) != 0) {
		printf("SetAdapter ioctl SIOCSIFADDR error is %d\n", ret);
		goto failed;
	}
	printf("SetAdapter Mask is %s\n", pAdapter->mask);
	if (ret = inet_aton(pAdapter->mask, &(sin->sin_addr)) == 0) {
		printf("SetAdapter inet_aton error is %d\n", ret);
		goto failed;
	}
	if (ret = ioctl(fd, SIOCSIFNETMASK, &ifr) != 0) {
		printf("SetAdapter ioctl SIOCSIFNETMAST error is %d\n", ret);
		goto failed;
	}

	memset(sin, 0, sizeof(struct sockaddr_in));
	sin->sin_family = AF_INET;
	sin->sin_port = 0;
	printf("SetAdapter Gateway is %s\n", pAdapter->gateway);
	if (ret = inet_aton(pAdapter->gateway, &(sin->sin_addr)) == 0) {
		printf("SetAdapter inet_aton error is %d\n", ret);
		goto failed;
	}
	memcpy(&rt.rt_gateway, sin, sizeof(struct sockaddr_in));
	((struct sockaddr_in *)&rt.rt_dst)->sin_family = AF_INET;
	((struct sockaddr_in *)&rt.rt_genmask)->sin_family = AF_INET;
	printf("flags is %d\n", RTF_GATEWAY);
	rt.rt_flags = RTF_GATEWAY;
	if (ret = ioctl(fd, SIOCADDRT, &rt) != 0) {
		printf("SetAdapter ioctl SIOCADDRT error is %d\n", ret);
		goto failed;
	}

	close(fd);
	return NET_ERROR_SUCCESS;
failed:
	close(fd);
	return NET_ERROR_FAILED;
}

#endif

