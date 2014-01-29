#= require "../components/jquery/jquery.js"

PackageStore =
  config:
    baseUrl: '/fixtures/list.json'
  init: (who) ->
    console.log who
    $.ajax
      url: @config.baseUrl
      success: (r) ->
        console.log r

$ ->
  PackageStore.init()
