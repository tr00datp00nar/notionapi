# About notionapi

This is an unofficial, Go API for https://notion.so. Mostly for reading, limited write capabilities.

It allows you to retrieve content of a Notion page in structured format.

You can then e.g. convert that format to HTML.

Note: official Notion API is still in beta and not as capable as this unofficial API.

Documentation:

- tutorial: https://blog.kowalczyk.info/article/c9df78cbeaae4e0cb2848c9964bcfc94/using-notion-api-go-client.html
- API docs: https://pkg.go.dev/github.com/kjk/notionapi

You can learn how [I reverse-engineered the Notion API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html) in order to write this library.

# Real-life usage

I use this API to publish my [blog](https://blog.kowalczyk.info/) and series of [programming books](https://www.programming-books.io/) from content stored in Notion.

Notion serves as a CMS (Content Management System). I write and edit pages in Notion.

I use custom Go program to download Notion pages using this this library and converts pages to HTML. It then publishes the result to Netlify.

You can see the code at https://github.com/kjk/blog and https://github.com/essentialbooks/tools/

# Implementations for other languages

- https://github.com/jamalex/notion-py : library for Python
- https://github.com/petersamokhin/knotion-api : library for Kotlin / Java
- https://github.com/Nishan-Open-Source/Nishan : library for node.js, written in Typescript

