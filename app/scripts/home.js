jQuery(function($) {
    var templateEngine = {
        compile: function(template) {
            var compiledTemplate = window.tmpl(template);
            return {
                render: compiledTemplate
            }
        }
    };

    $("#search")
        .typeahead({
            name: "artifacts",
            // remote: "/search.json?q=%QUERY",
            local: [
                {
                    tokens: ["guava", "com.google.guava:guava" ],
                    a: "guava",
                    g: "com.google.guava",
                    l: "16.0-rc1",
                    value: "com.google.guava:guava",
                    vc: "31"
                },
            ],
            template: $("script#typeahead-template").html(),
            engine: templateEngine
        })
        .on("typeahead:selected", function(ev, datum) {
            window.location = "/artifact/" + datum.g + "/" + datum.a;
        });
});
