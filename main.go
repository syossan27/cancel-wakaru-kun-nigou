package main

import (
  "github.com/ant0ine/go-json-rest/rest"
  "github.com/PuerkitoBio/goquery"
  "github.com/k0kubun/pp"
  "log"
  //"fmt"
  "net/http"
  "sync"
)

type PostData struct {
  Url string
}

type List struct {
  Url []string
  User []User
}

type User struct {
  Name string
  Image string
  CancelCount int
  JoinCount int
}

func main() {
  api := rest.NewApi()
  api.Use(rest.DefaultDevStack...)
  router, err := rest.MakeRouter(
    rest.Post("/cancel", PostCancel),
  )

  if err != nil {
    log.Fatal(err)
  }

  api.SetApp(router)
  log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func PostCancel(w rest.ResponseWriter, r *rest.Request) {
  post_data := PostData{}
  err := r.DecodeJsonPayload(&post_data)
  if err != nil {
    rest.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if post_data.Url == "" {
    rest.Error(w, "url required", 400)
  }

  list := List{}
  GetPageToConnpass(post_data.Url, &list)

  wg := new(sync.WaitGroup)
  for _, url := range list.Url {
    wg.Add(1)
    go GetUserPageToConnpass(&list, url, wg)
  }
  wg.Wait()
  pp.Println(list.User)
}

func GetPageToConnpass(url string, list *List) {
  doc, _ := goquery.NewDocument(url + "participation/#participants")
  doc.Find(".user").Each(func(_ int, s *goquery.Selection) {
    s.Find(".image_link").Each(func(_ int, s *goquery.Selection) {
      url, _ := s.Attr("href")
      list.Url = append(list.Url, url)
    })
  })
}

func GetUserPageToConnpass(list *List, url string, wg *sync.WaitGroup) {
  user := User{"", "", 0, 0}

  doc, _ := goquery.NewDocument(url)
  user.Name, _ = doc.Find("#side_area > div.mb_20.text_center img").Attr("title")
  doc.Find("#main > div.event_area.mb_10 > div.event_list.vevent").Each(func(_ int, s *goquery.Selection) {
    join_status := s.Find("p.label_status_tag").Text()
    if join_status == "キャンセル" {
      user.CancelCount++
    } else {
      user.JoinCount++
    }
  })

  // TODO: 要実装
  // ページネーションが存在し、かつ最終ページではない場合に次のページを取得する
  if doc.Is("#main > div.paging_area > ul > li.active > span") && doc.Is("#main > div.paging_area > ul > li.active + li"){
  }

  list.User = append(list.User, user)
  wg.Done()
}
