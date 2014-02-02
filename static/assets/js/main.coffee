#= require "../components/jquery/jquery.js"

PackageStore =
  config:
    baseUrl: '/fixtures/list.json'

  process: (json) ->
    html = $('.box').html()
    $('.box').remove()
    for el in json['packages']
      box = $('<div class="box" />').html(html)
      box.find('h2').text el.importPath
      box.data('importPath', el.importPath)
      box.find('.message').html el.commitMessages.join('<br />')
      box.find('.avatar img').attr 'src', el.image
      box.find('.btn').on 'click', ->
        importPath = $(@).parents('.box').data('importPath')
        $.ajax
          url: '/-/post'
          data:
            path: importPath
          success: (e) ->
            console.log e
          error: (e) ->
            console.error "Error:"

      $('body').append box

  init: (who) ->

    $.ajax
      url: @config.baseUrl
      success: @process
$ ->
  PackageStore.init()
