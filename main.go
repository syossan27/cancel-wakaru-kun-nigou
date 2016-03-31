package main

import (
  "github.com/ant0ine/go-json-rest/rest"
  "github.com/PuerkitoBio/goquery"
  //"github.com/k0kubun/pp"
  "log"
  "fmt"
  "net/http"
  "sync"
  "strconv"
  "runtime"
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
    rest.Post("/", PostCancel),
  )

  if err != nil {
    log.Fatal(err)
  }

  api.SetApp(router)
  log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func PostCancel(w rest.ResponseWriter, r *rest.Request) {
  cpus := runtime.NumCPU()
  runtime.GOMAXPROCS(cpus)

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
  fmt.Println(post_data.Url)
  GetPageToConnpass(post_data.Url, &list)

  wg := new(sync.WaitGroup)
  for _, url := range list.Url {
    wg.Add(1)
    go GetUserPageToConnpass(&list, url, wg)
  }
  wg.Wait()

  w.WriteJson(list.User)
  // pp.Println(list.User)
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
  // 退会ユーザーなどはURLが取れないため無視
  if url != "" {
    user := User{"", "", 0, 0}

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

    list.User = append(list.User, user)
  }
  wg.Done()
}
