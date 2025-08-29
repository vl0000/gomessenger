# GoMessenger

GoMessenger is a web-based messaging app built with ConnectRPC.

![](./doc/assets/demonstration.GIF)
explain the public folder
link to the front end repo

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
