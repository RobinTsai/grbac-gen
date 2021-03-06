package parser

import (
	"encoding/json"
	"fmt"
	"grbac-gen/pkg/utils"
	"log"
	"regexp"
	"strings"
)

type PermissionDoc struct {
	Id              int      `json:"id"`
	Host            string   `json:"host"`             // default *
	Path            string   `json:"path"`             // default *
	Method          string   `json:"method"`           // "{GET}"
	AuthorizedRoles []string `json:"authorized_roles"` //
	ForbiddenRoles  []string `json:"forbidden_roles"`
	AllowAnyone     bool     `json:"allow_anyone"`

	Methods               []string `json:"-"` // 临时使用
	PermKey               string   `json:"-"` // +acc_roles-deny_roles
	Frags                 []string `json:"-"` //
	SameFragCountWithLast int      `json:"-"` // if key 不一样 0
}

func (p *PermissionDoc) GetMethodsFromMethodStr() []string {
	s := strings.TrimPrefix(p.Method, "{")
	s = strings.TrimSuffix(s, "}")
	return strings.Split(s, ",")
}
func (p *PermissionDoc) SetMethodFromMethods() {
	p.Methods = utils.UniqueStrings(p.Methods, false)
	s := strings.Join(p.Methods, ",")
	p.Method = "{" + s + "}"
}

type Permission struct {
	*PermissionDoc
	Pkg                   string   `json:"pkg"`
	Filepath              string   `json:"filepath"`
	RawRouterLine         string   `json:"rawRouterLine"`
	RawAuthRolesLine      string   `json:"rawAuthRolesLine"`
	RawForbiddenRolesLine string   `json:"rawForbiddenRolesLine"`
	Tags                  []string `json:"-"`
}

func (p *Permission) Parse() error {
	if p.RawRouterLine == "" {
		log.Println("empty permission raw ")
		return fmt.Errorf("empty permission raw router line")
	}

	if err := p.parseRouterLine(); err != nil {
		return err
	}

	p.AuthorizedRoles = p.parseRolesLine(p.RawAuthRolesLine, "*")
	p.ForbiddenRoles = p.parseRolesLine(p.RawForbiddenRolesLine, "")

	return nil
}

// @Router       /admin/users [get]
var routerPattern = regexp.MustCompile(`^(/[\w./\-{}+:$]*)[[:blank:]]+\[(\w+)]`)

var routerRegex = regexp.MustCompile(`\{[^{]*}`)

func (p *Permission) parseRouterLine() error {
	matches := routerPattern.FindStringSubmatch(p.RawRouterLine)
	if len(matches) != 3 {
		return fmt.Errorf("can not parse router comment \"%s\"", p.RawRouterLine)
	}

	p.Path = string(routerRegex.ReplaceAll([]byte(matches[1]), []byte("*")))
	p.Method = strings.ToUpper(matches[2])

	return nil
}

func (p *Permission) parseRolesLine(line string, def string) []string {
	res := make([]string, 0, 3)
	arr := strings.Split(line, ",")
	for _, s := range arr {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		res = append(res, s)
	}
	if def != "" && len(res) == 0 {
		res = []string{def}
	}
	return res
}

func (p *Permission) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.PermissionDoc)
}
