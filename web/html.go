package web

const (
	html_ = `
  <!DOCTYPE html>
  <html lang="ru" data-theme="light">
    <head>
      <title>%s</title>
      <meta charset=utf-8>
      <meta http-equiv=X-UA-Compatible content='IE=edge;chrome=1'>
      <meta name=viewport content="width=device-width, initial-scale=1">
      <link rel="apple-touch-icon" sizes="180x180" href="favicon/apple-touch-icon.png">
      <link rel="icon" type="image/svg+xml" sizes="32x32" href="favicon/doge.svg">
      <link rel="icon" type="image/png" sizes="16x16" href="favicon/favicon-16x16.png">
      <link rel="manifest" href="favicon/site.webmanifest">
      <link rel="mask-icon" href="favicon/safari-pinned-tab.svg" color="#5bbad5">
      <meta name="theme-color" media="(prefers-color-scheme: light)" content="#fefbf4">
      <meta name="theme-color" media="(prefers-color-scheme: dark)" content="#444">
      <link href="m/doge.css" rel="stylesheet" type="text/css" />
      <script src="m/doge.js"></script>
    </head>
    <body>
      <!-- Header -->
      <header >
          <!-- Navigation -->
          <nav >
          <ul>
              <li>
                  <a href="/"><img style="width:2em; height: 2em; margin-top: -0.5em;" src="favicon/doge.svg"/> </a>
              </li>
              <li><a href="/">doge &middot; news</a></li>
              <li class="float-right sticky"><a onclick="addFontSize(-1)">ᴀ-</a>|<a onclick="addFontSize(1)">A+</a></li>
              <li class="float-right sticky"><a onclick="toggleDarkMode(this)">☪</a></li>
              <li ><a href="#basic">period ▾</a>
              <ul>
                  <li><a href="/">now</a></li>
                  <li><a href="td">today</a></li>
                  <li><a href="ytd">yesterday</a></li>
                  <li><a href="wk">week</a></li>
              </ul>
              </li>
              <li class="float-right"><a href="https://github.com/recoilme/dogenews">@github</a></li>
              <li class="float-right">
                  <script async src="https://telegram.org/js/telegram-widget.js?15" data-telegram-login="newsdogebot" data-size="medium" data-radius="4" data-auth-url="https://doge.news/auth" data-request-access="write"></script>
              </li>
              </ul>
          </nav>
      </header>
      <!-- Main page -->
      <main>
        %s
      </main> 
      <footer>
        <hr>
      </footer>
    </body>
  </html>
  `

	article_ = `
  <article>
    <section> 
        <h2><a href="%s">%s</a></h2>
        <p>
          %s
        </p>
        <time>%s</time>
        <div class="float-right"><code>%s</code>&ensp;%s%s</div>
    </section>
  </article>  
  `
)
