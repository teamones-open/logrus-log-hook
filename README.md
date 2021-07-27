# logrus-log-hook
teamones log 服务上报 logrus hook

# install

go get github.com/teamones-open/logrus-log-hook

## usage
```go
	hook := httphook.New(
		"service-name",
		"localhost:8081/monolog/add",
		logrus.AllLevels,
	)

	hook.BeforePost = func(req *http.Request) error {
		return nil
	}
	hook.AfterPost = func(res *http.Response) error {
		return nil
	}

	logrus.AddHook(hook)
```