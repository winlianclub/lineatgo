package lineatgo

import (
    "net/http"
    "fmt"
    "io/ioutil"
    "strings"
    "net/url"
    "github.com/PuerkitoBio/goquery"
    "github.com/pkg/errors"
)

const (
    Administrator = "ADMIN"
    Operator = "OPERATOR"
    LimitedOperator = "OPERATOR_LIMITED"
    Messenger = "MESSENGER"
)
/*
GetAuthURL retrieve a url to enable access the account.
 */
func (b *Bot) GetAuthURL(role string) string {
    v := url.Values{"role": {role}}
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/auth/url", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.Api.XRT)
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    return string(cont)
}

type AuthUserList struct {
    Users []AuthUser
}

type AuthUser struct {
    Name string
    Id string
    BotId string
    IsPaymaster bool
    AuthorityType string
    Api *Api
}

func (b *Bot) findAuthUser() {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/userlist/", b.BotId), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.Api.XRT)
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()
    var ul []AuthUser
    doc, _ := goquery.NewDocumentFromResponse(resp)
    doc.Find("div.MdCMN08ImgSet").Each(func(_ int, s *goquery.Selection) {
        t := s.Find("p.mdCMN08Ttl").Text()
        u := parseAuthTxt(t)
        u.Api = b.Api
        u.BotId = b.BotId
        imgurl, _ := s.Find("div.mdCMN08Img > img").Attr("src")
        u.Id = imgurl[len(fmt.Sprintf("/%v/userlist/profile/", b.BotId)):]
        ul = append(ul, u)
    })
    b.AuthUserList = &AuthUserList{Users: ul}
}

func parseAuthTxt(t string) AuthUser {
    var u AuthUser
    if strings.Contains(t, "Paymaster") {
        u.IsPaymaster = true
    }
    if strings.Contains(t, "Administrator") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[13 + addition:]
        u.AuthorityType = "Administrator"
    }
    if strings.Contains(t, "Operations personnel (no statistics view)") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[41 + addition:]
        u.AuthorityType = "Operator(no statistics view)"
    }
    if strings.Contains(t, "Operations personnel (no authority to send)") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[43 + addition:]
        u.AuthorityType = "Operator(no authority to send)"
    }
    return u
}

/*
DeleteAuthUser eliminate authenticated user
 */
func (u *AuthUser) Delete() error {
    if u.IsPaymaster {
        return errors.New("ERROR: This user is a paymaster. Please execute SetPaymaster to other user.")
    }
    delurl := fmt.Sprintf("/%v/userlist/del/%v",u.BotId,  u.Id)
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me%v", delurl), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", u.Api.XRT)
    resp, _ := u.Api.Client.Do(request)
    defer resp.Body.Close()
    return nil
}

func (u AuthUser) SetPaymaster()  {
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/api/users/payperson/%v", u.BotId, u.Id), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", u.Api.XRT)
    resp, _ := u.Api.Client.Do(request)
    defer resp.Body.Close()
}