package main

import (
  "log"
  "net/http"
  "strconv"
  "os"
  // "fmt"

  "github.com/ant0ine/go-json-rest/rest"
  "github.com/PuerkitoBio/goquery"
  "github.com/k0kubun/pp"
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
    rest.Post("/join", PostJoin),
    rest.Post("/cancel", PostCancel),
  )

  api.Use(&rest.CorsMiddleware{
    RejectNonCorsRequests: false,
    OriginValidator: func(origin string, request *rest.Request) bool {
      // allow every origin (for now)
      return true
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{
      "Accept", "Content-Type", "X-Custom-Header", "Origin"},
      AccessControlAllowCredentials: true,
      AccessControlMaxAge:           3600,
    }
  )

  if err != nil {
    log.Fatal(err)
  }

  port := os.Getenv("PORT")

  api.SetApp(router)
  log.Fatal(http.ListenAndServe(":" + port, api.MakeHandler()))
  // log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func PostJoin(w rest.ResponseWriter, r *rest.Request) {
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
  w.WriteJson(list.Url)
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

  url := post_data.Url

  user := GetUserPageToConnpass(url)
  pp.Println(user)

  w.WriteJson(user)
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

func GetUserPageToConnpass(url string) User {
  user := User{"", "", 0, 0}

  // 退会ユーザーなどはURLが取れないため無視
  if url != "" {
    doc, _ := goquery.NewDocument(url)
    image_elm := doc.Find("#side_area > div.mb_20.text_center img")
    user.Name, _ = image_elm.Attr("title")
    user.Image, _ = image_elm.Attr("src")
    doc.Find("#main > div.event_area.mb_10 > div.event_list.vevent").Each(func(_ int, s *goquery.Selection) {
      join_status := s.Find("p.label_status_tag").Text()
      if join_status == "キャンセル" {
        user.CancelCount++
      } else {
        user.JoinCount++
      }
    })

    // ページ数が１以上ある場合
    if (doc.Find("#main > div.paging_area > ul > li").Length() - 1) > 1 {
      total_page := doc.Find("#main > div.paging_area > ul > li").Length() - 1

      for i := 2; i <= total_page; i++ {
        doc, _ := goquery.NewDocument(url + "?page=" + strconv.Itoa(i))
        doc.Find("#main > div.event_area.mb_10 > div.event_list.vevent").Each(func(_ int, s *goquery.Selection) {
          join_status := s.Find("p.label_status_tag").Text()
          if join_status == "キャンセル" {
            user.CancelCount++
          } else {
            user.JoinCount++
          }
        })
      }
    }
  }

  return user
}
