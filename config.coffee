exports.config =
  modules:
    definition: false
    wrapper: false
  files:
    javascripts:
      joinTo:
        'js/vendor.js': /^bower_components|vendor/
        'js/app.js': /^app\/scripts/
    stylesheets:
      joinTo:
        'css/app.css': /^(app|bower_components)/
  plugins:
    imageoptimizer:
      smushit: false
      path: "images"