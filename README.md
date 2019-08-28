# vortices

Vortices is an organizer to test ICE and WebRTC under different network
environments.

WebRTC, and its underlying ICE method, depend on being able to create a channel
between two peers where both can send UDP packages to one another. This is
not easy to test as it depends on the network environment each is located,
and how their NAT works.

## Dependencies

```bash
$ sudo apt install docker-compose
```
