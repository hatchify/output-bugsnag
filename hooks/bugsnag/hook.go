package bugsnag

import (
	"fmt"
	"os"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/hatchify/output/stackcache"
	"github.com/sirupsen/logrus"
)

// HookOptions allows to set additional Hook options.
type HookOptions struct {
	// Levels enables this hook for all listed levels.
	Levels       []logrus.Level

	Env               string
	AppVersion        string
	BugsnagAPIKey     string
	BugsnagEnabledEnv []string
	BugsnagPackages   []string
}

func checkHookOptions(opt *HookOptions) *HookOptions {
	if opt == nil {
		opt = &HookOptions{}
	}
	if len(opt.Levels) == 0 {
		opt.Levels = []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}
	}
	if len(opt.Env) == 0 {
		opt.Env = os.Getenv("OUTPUT_ENV")
		if len(opt.Env) == 0 {
			opt.Env = "local"
		}
	}
	if len(opt.AppVersion) == 0 {
		opt.AppVersion = os.Getenv("OUTPUT_APP_VERSION")
	}
	if len(opt.BugsnagAPIKey) == 0 {
		opt.BugsnagAPIKey = os.Getenv("OUTPUT_BUGSNAG_KEY")
	}
	if len(opt.BugsnagEnabledEnv) == 0 {
		opt.BugsnagEnabledEnv = []string{
			"prod",
			"staging",
			"test",
		}
	}
	if len(opt.BugsnagPackages) == 0 {
		opt.BugsnagPackages = []string{
			"main",
			"github.com/Hatch1fy/*",
			"github.com/hatchify/*",
		}
	}
	return opt
}

// NewHook initializes a new logrus.Hook using provided params and options.
func NewHook(opt *HookOptions) logrus.Hook {
	opt = checkHookOptions(opt)
	internalLogger := logrus.New()
	internalLogger.SetLevel(logrus.ErrorLevel)
	return &hook{
		opt:   opt,
		stack: stackcache.New(6, "github.com/hatchify/output"),
		notifier: bugsnag.New(bugsnag.Configuration{
			APIKey:              opt.BugsnagAPIKey,
			ReleaseStage:        opt.Env,
			ProjectPackages:     opt.BugsnagPackages,
			AppVersion:          opt.AppVersion,
			NotifyReleaseStages: opt.BugsnagEnabledEnv,
			PanicHandler:        func() {},
			Logger:              internalLogger.WithField("package", "bugsnag"),
		}),
	}
}

type hook struct {
	opt      *HookOptions
	stack    stackcache.StackCache
	notifier *bugsnag.Notifier
}

func (h *hook) Levels() []logrus.Level {
	return h.opt.Levels
}

func (h *hook) Fire(e *logrus.Entry) error {
	var err ErrorWithStackFrames
	var errContext bugsnag.Context
	// check if we have error in fields
	if withErr, ok := e.Data["error"].(error); ok {
		// check if that error has stack (was wrapped at some point)
		if withStack, ok := withErr.(ErrorWithStackFrames); ok {
			// use this error to report, with its original stack
			err = withStack
			errContext.String = e.Message
		} else {
			// no stack with error, wrap it
			stackFrames := h.stack.GetStackFrames()
			err = newErrorWithStackFrames(withErr, stackFrames)
			errContext.String = e.Message
		}
	} else {
		// no error within fields, construct new one from log message
		stackFrames := h.stack.GetStackFrames()
		err = newErrorWithStackFrames(fmt.Errorf("%s", e.Message), stackFrames)
	}
	needSync := false
	severity := bugsnag.SeverityInfo
	switch e.Level {
	case logrus.WarnLevel:
		severity = bugsnag.SeverityWarning
	case logrus.ErrorLevel:
		severity = bugsnag.SeverityError
	case logrus.FatalLevel, logrus.PanicLevel:
		severity = bugsnag.SeverityError
		needSync = true
	}
	user := bugsnag.User{}
	if userID, ok := e.Data["@user.id"].(string); ok {
		user.Id = userID
		delete(e.Data, "@user.id")
	}
	if userName, ok := e.Data["@user.name"].(string); ok {
		user.Name = userName
		delete(e.Data, "@user.name")
	}
	if userEmail, ok := e.Data["@user.email"].(string); ok {
		user.Email = userEmail
		delete(e.Data, "@user.email")
	}
	metaData := fieldsToMetaData(e.Data)
	if len(errContext.String) > 0 {
		_ = h.notifier.NotifySync(err, needSync, severity, metaData, user, errContext)
		return nil
	}
	_ = h.notifier.NotifySync(err, needSync, severity, metaData, user)
	return nil
}

func fieldsToMetaData(fields logrus.Fields) bugsnag.MetaData {
	if len(fields) == 0 {
		return bugsnag.MetaData{}
	}
	fieldsMap := make(map[string]interface{}, len(fields))
	for field, value := range fields {
		switch field {
		case "blob", "error", "@user.id", "@user.name", "@user.email":
			continue
		}
		fieldsMap[field] = value
	}
	return bugsnag.MetaData{
		"Fields": fieldsMap,
	}
}
