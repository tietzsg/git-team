package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	addExecutor "github.com/hekmekk/git-team/src/add"
	"github.com/hekmekk/git-team/src/config"
	execDisable "github.com/hekmekk/git-team/src/disable"
	enableExecutor "github.com/hekmekk/git-team/src/enable"
	git "github.com/hekmekk/git-team/src/gitconfig"
	removeExecutor "github.com/hekmekk/git-team/src/remove"
	statusApi "github.com/hekmekk/git-team/src/status"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	version = "v1.1.2"
	author  = "Rea Sand <hekmek@posteo.de>"
)

type add struct {
	command  *kingpin.CmdClause
	alias    *string
	coauthor *string
}

func newAdd(app *kingpin.Application) add {
	command := app.Command("add", "Add an alias")
	return add{
		command:  command,
		alias:    command.Arg("alias", "The alias to be added").Required().String(),
		coauthor: command.Arg("coauthor", "The co-author").Required().String(),
	}
}

type remove struct {
	command *kingpin.CmdClause
	alias   *string
}

func newRemove(app *kingpin.Application) remove {
	command := app.Command("rm", "Remove an alias")
	return remove{
		command: command,
		alias:   command.Arg("alias", "The alias to be removed").Required().String(),
	}
}

type enable struct {
	command             *kingpin.CmdClause
	aliasesAndCoauthors *[]string // can contain both aliases and coauthors
}

func newEnable(app *kingpin.Application) enable {
	command := app.Command("enable", "Provisions a git-commit template with the provided co-authors. A co-author must either be an alias or of the shape \"Name <email>\"").Default()
	return enable{
		command:             command,
		aliasesAndCoauthors: command.Arg("coauthors", "Git co-authors").Strings(),
	}
}

type disable struct {
	command *kingpin.CmdClause
}

func newDisable(app *kingpin.Application) disable {
	return disable{
		command: app.Command("disable", "Use default template"),
	}
}

type status struct {
	command *kingpin.CmdClause
}

func newStatus(app *kingpin.Application) status {
	return status{
		command: app.Command("status", "Print the current status"),
	}
}

type list struct {
	command *kingpin.CmdClause
}

func newList(app *kingpin.Application) list {
	command := app.Command("list", "List currently available aliases")
	command.Alias("ls")
	return list{
		command: command,
	}
}

type application struct {
	app     *kingpin.Application
	add     add
	remove  remove
	enable  enable
	disable disable
	status  status
	list    list
}

func newApplication() application {
	app := kingpin.New("git-team", "Command line interface for creating git commit templates provisioned with one or more co-authors. Please note that \"git commit -m\" is not affected by commit templates.")

	app.HelpFlag.Short('h')
	app.Version(version)
	app.Author(author)

	return application{
		app:     app,
		add:     newAdd(app),
		remove:  newRemove(app),
		enable:  newEnable(app),
		disable: newDisable(app),
		status:  newStatus(app),
		list:    newList(app),
	}
}

func runEnable(application application) {
	enableDeps := enableExecutor.Dependencies{
		CreateDir:         os.MkdirAll,           // TODO: CreateTemplateDir
		WriteFile:         ioutil.WriteFile,      // TODO: WriteTemplateFile
		SetCommitTemplate: git.SetCommitTemplate, // TODO: GitSetCommitTemplate
		GitResolveAliases: git.ResolveAliases,
		PersistEnabled:    statusApi.PersistEnabled,
		LoadConfig:        config.Load,
	}
	execEnable := enableExecutor.ExecutorFactory(enableDeps)
	cmd := enableExecutor.Command{
		Coauthors: append(*application.enable.aliasesAndCoauthors),
	}
	enableErrs := execEnable(cmd)
	exitIfErr(enableErrs...)

	status, err := statusApi.Fetch()
	exitIfErr(err)

	fmt.Println(status.ToString())
	os.Exit(0)
}

func runDisable(application application) {
	err := execDisable.Exec()
	exitIfErr(err)

	status, err := statusApi.Fetch()
	exitIfErr(err)

	fmt.Println(status.ToString())
	os.Exit(0)
}

func runStatus(application application) {
	status, err := statusApi.Fetch()
	exitIfErr(err)

	fmt.Println(status.ToString())
	os.Exit(0)
}

func runAdd(application application) {
	addDeps := addExecutor.Dependencies{
		AddGitAlias: git.AddAlias,
	}
	execAdd := addExecutor.ExecutorFactory(addDeps)

	addAlias := *application.add.alias
	addCoauthor := *application.add.coauthor

	cmd := addExecutor.Command{
		Alias:    addAlias,
		Coauthor: addCoauthor,
	}
	addErr := execAdd(cmd)
	exitIfErr(addErr)

	color.Green(fmt.Sprintf("Alias '%s' -> '%s' has been added.", addAlias, addCoauthor))
	os.Exit(0)
}

func runRemove(application application) {
	rmDeps := removeExecutor.Dependencies{
		GitResolveAlias: git.ResolveAlias,
		GitRemoveAlias:  git.RemoveAlias,
	}
	execRemove := removeExecutor.ExecutorFactory(rmDeps)

	removeAlias := *application.remove.alias

	cmd := removeExecutor.Command{
		Alias: removeAlias,
	}

	rmErr := execRemove(cmd)
	exitIfErr(rmErr)

	color.Red(fmt.Sprintf("Alias '%s' has been removed.", removeAlias))
	os.Exit(0)
}

func runList(application application) {
	assignments := git.GetAddedAliases() // TODO: git.GetAssignments()

	blackBold := color.New(color.FgBlack).Add(color.Bold)
	blackBold.Println("Aliases:")
	blackBold.Println("--------")

	for alias, coauthor := range assignments {
		color.Magenta(fmt.Sprintf("'%s' -> '%s'", alias, coauthor))
	}
	os.Exit(0)
}

func main() {
	application := newApplication()

	switch kingpin.MustParse(application.app.Parse(os.Args[1:])) {
	case application.enable.command.FullCommand():
		runEnable(application)
	case application.disable.command.FullCommand():
		runDisable(application)
	case application.status.command.FullCommand():
		runStatus(application)
	case application.add.command.FullCommand():
		runAdd(application)
	case application.remove.command.FullCommand():
		runRemove(application)
	case application.list.command.FullCommand():
		runList(application)
	}
}

func exitIfErr(validationErrs ...error) {
	if len(validationErrs) > 0 && validationErrs[0] != nil {
		os.Stderr.WriteString(fmt.Sprintf("error: %s\n", foldErrors(validationErrs)))
		os.Exit(-1)
	}
}

func foldErrors(validationErrors []error) error {
	var buffer bytes.Buffer
	for _, err := range validationErrors {
		buffer.WriteString(err.Error())
		buffer.WriteString("; ")
	}
	return errors.New(strings.TrimRight(buffer.String(), "; "))
}
