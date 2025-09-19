# GoMessenger

GoMessenger is a web-based messaging app built with ConnectRPC created for learning purposes.

![](./doc/assets/demonstration.GIF)

[Front-end repo](https://github.com/vl0000/gomessenger-frontend)
## Documentation

[Data flow model](https://github.com/vl0000/gomessenger/blob/main/doc/Dfd.md)

[Design document](https://github.com/vl0000/gomessenger/blob/main/doc/MessengerApp.md)

## Building it with docker
Add your secret key to the `SECRET_KEY` enviroment variable. If the key is not set, JWTs **will not work**.
Afterwards, open a console and simply use:
```bash
docker build . -t [tag of your choice]
```
To run:
```Bash
docker run -p 3000:80 [image_id]
```
When building with podman, add the flag `--format docker`.
