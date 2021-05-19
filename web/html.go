package web

const (
	menu = `
<header class="text-gray-400 bg-gray-900 body-font">
<div class="container mx-auto flex flex-wrap p-5 flex-col md:flex-row items-center">
<a href="/" class="flex title-font font-medium items-center text-white mb-4 md:mb-0">
  <img class="w-10 h-10 text-white p-2 bg-green-500 rounded-full" viewBox="0 0 24 24" src="web/favicon/doge.svg"/>
  <span class="ml-3 text-xl">doge &middot; news</span>
</a>
<nav class="md:mr-auto md:ml-4 md:py-1 md:pl-4 md:border-l md:border-gray-700	flex flex-wrap items-center text-base justify-center">
  <a href="today" class="mr-5 hover:text-white">Today</a>
  <a href="yesterday" class="mr-5 hover:text-white">Yesterday</a>
  <a href="week" class="mr-5 hover:text-white">Week</a>
</nav>
<!--
<button class="inline-flex items-center bg-gray-800 border-0 py-1 px-3 focus:outline-none hover:bg-gray-700 rounded text-base mt-4 md:mt-0">Button
  <svg fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" class="w-4 h-4 ml-1" viewBox="0 0 24 24">
	<path d="M5 12h14M12 5l7 7-7 7"></path>
  </svg>
</button>
-->
<script async src="https://telegram.org/js/telegram-widget.js?15" data-telegram-login="newsdogebot" data-size="medium" data-radius="4" data-auth-url="https://doge.news/auth" data-request-access="write"></script>
</div>
</header>
`

	artHead = `
<section class="text-gray-400 bg-gray-900 body-font overflow-hidden">
  <div class="container px-5 py-24 mx-auto">
	<div class="flex flex-wrap -m-12">
`
	artFoot = `
	</div>
  </div>
</section>
`

	artBody = `
<div class="p-12 md:w-1/2 flex flex-col items-start">
<span class="inline-block py-1 px-2 rounded bg-gray-800 text-gray-400 text-opacity-75 text-xs font-medium tracking-widest">%s</span>

<h1 class="title-font sm:text-2xl text-xl font-medium text-white mb-3">%s</h1>
<p class="leading-relaxed mb-8">%s</p>
<div class="flex items-center flex-wrap pb-4 mb-4 border-b-2 border-gray-800 border-opacity-75 mt-auto w-full">
  <a href="%s" class="text-green-400 inline-flex items-center">More
	<svg class="w-4 h-4 ml-2" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
	  <path d="M5 12h14"></path>
	  <path d="M12 5l7 7-7 7"></path>
	</svg>
  </a>
  <span class="text-gray-500 mr-3 inline-flex items-center ml-auto leading-none text-sm pr-3 py-1 border-r-2 border-gray-800">
	<!--<svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
	  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
	  <circle cx="12" cy="12" r="3"></circle>
	</svg>-->%s
  </span>
  <span class="text-gray-500 mr-3 inline-flex items-center leading-none text-sm pr-3 py-1 border-r-2 border-gray-800">
	<svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
	  <path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z"></path>
	</svg>%s
  </span>
  <span class="text-gray-500 inline-flex items-center leading-none text-sm pr-3 py-1 ">
  <svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
		<path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"></path>
	</svg>%s
  </span>
</div>
<a class="inline-flex items-center">
  <img alt="blog" src="%s" class="w-12 h-12 rounded-full flex-shrink-0 object-cover object-center">
  <span class="flex-grow flex flex-col pl-4">
	<span class="title-font font-medium text-white">%s</span>
	<span class="text-gray-500 text-xs tracking-widest mt-0.5">%s</span>
  </span>
</a>
</div>

`

	htmlHead = `
<!DOCTYPE html>
<html lang=ru>

<head>
	<meta charset=utf-8>
	<meta http-equiv=X-UA-Compatible content='IE=edge;chrome=1'>
	<meta name=viewport content="width=device-width, initial-scale=1">
	<title>doge &middot; news</title>
    <link rel="apple-touch-icon" sizes="180x180" href="favicon/apple-touch-icon.png">
    <link rel="icon" type="image/svg+xml" sizes="32x32" href="favicon/doge.svg">
    <link rel="icon" type="image/png" sizes="16x16" href="favicon/favicon-16x16.png">
    <link rel="manifest" href="favicon/site.webmanifest">
    <link rel="mask-icon" href="favicon/safari-pinned-tab.svg" color="#5bbad5">
    <meta name="msapplication-TileColor" content="#da532c">
    <meta name="theme-color" content="#ffffff">
    <meta name=msapplication-TileColor content="#2b5797">
	<meta name=theme-color content="#007bff">
</head>
`

	artHero = `
    <!--<div class="flex flex-wrap -m-4">-->
<div class="p-9 md:w-1/2">
  <div class="h-full border-2 border-gray-800 rounded-lg overflow-hidden">
    <img class="lg:h-48 md:h-36 w-full object-cover object-center" src="%s" alt="blog">
    <div class="p-6">
      <h2 class="tracking-widest text-xs title-font font-medium text-gray-500 mb-1">%s</h2>
      <h1 class="title-font text-lg font-medium text-white mb-3">%s</h1>
      <p class="leading-relaxed mb-3">%s</p>
      <div class="flex items-center flex-wrap ">
        <a href="%s" class="text-green-400 inline-flex items-center md:mb-2 lg:mb-0">More
          <svg class="w-4 h-4 ml-2" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 12h14"></path>
            <path d="M12 5l7 7-7 7"></path>
          </svg>
        </a>
        <span class="text-gray-500 mr-3 inline-flex items-center lg:ml-auto md:ml-0 ml-auto leading-none text-sm pr-3 py-1 border-r-2 border-gray-800">
          <svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>%s
        </span>
        <span class="text-gray-500 inline-flex items-center leading-none text-sm">
          <svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
            <path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z"></path>
          </svg>%s
        </span>
      </div>
    </div>
  </div>
</div>
`

	artHero2 = `
<div class="p-4 lg:w-1/2">
  <div class="h-full bg-gray-800 bg-opacity-40 px-8 pt-16 pb-24 rounded-lg overflow-hidden text-center relative">
  
    <h2 class="tracking-widest text-xs title-font font-medium text-gray-500 mb-1">%s</h2>
    <h1 class="title-font sm:text-2xl text-xl font-medium text-white mb-3">%s</h1>
    <p class="leading-relaxed mb-3">%s</p>

    <a class="text-green-400 inline-flex items-center" href="%s">More
      <svg class="w-4 h-4 ml-2" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round">
        <path d="M5 12h14"></path>
        <path d="M12 5l7 7-7 7"></path>
      </svg>
    </a>

    <div class="text-center mt-2 leading-none flex flex-col justify-center absolute bottom-0 left-0 w-full py-4">

      <div class="py-2">
        <span class="title-font font-medium text-white ">%s</span>
      </div>
      <div>  
        <span class="text-gray-500 mr-3 inline-flex items-center leading-none text-sm pr-3 py-1 border-r-2 border-gray-700 border-opacity-50">
          <svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
            <path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z"></path>
          </svg>%s
        </span>
        <span class="text-gray-500 inline-flex items-center leading-none text-sm">
          <svg class="w-4 h-4 mr-1" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
		        <path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"></path>
	        </svg>%s
        </span>
      </div>
    </div>
  </div>
</div>
`
)
