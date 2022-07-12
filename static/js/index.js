// path and named for label
const supportedAlgos = {
  "/ctph": "ssdeep",
};
const algoPaths = Object.keys(supportedAlgos);

// Wires in the algorithm implementation selected to load
// https://stackoverflow.com/a/39695533/1120453
function fetchApp() {
  var algo = () => {return localStorage.getItem("algoPathName")};
  if (null === algo() || !algoPaths.includes(algo())){
    // set to DEFAULT
    localStorage.setItem("algoPathName", algoPaths[0]);
  }
  $("#algo-name").text(supportedAlgos[algo()]);

  $.ajax({
    url: "fragments/app.html",
    type: "GET",
    cache: false,
    dataType: "html",
  })
  .fail(function (error) {
    console.log("ajax failed: ", error);
  })
  .done(function (response) {
    $("#app").html(response);
  });
}

$(document).ready((e) => {
  $(".navbar-item").click((e) => {
    if (e.target.matches("[data-link]")) {
      e.preventDefault();
      localStorage.setItem("algoPathName", e.target.pathname);
      fetchApp();
    }
  });

  fetchApp();
});

// Bulma toggle for mobile hamburger menu
$(document).ready(() => {
  $(".navbar-burger").click(() => {
    $(".navbar-burger").toggleClass("is-active");
    $(".navbar-menu").toggleClass("is-active");
  });
});

function syncInputAndOutputWidths() {
  $("#algo-output").css("width", $("#algo-input").css("width"));
}

$(document).ready(() => {
  syncInputAndOutputWidths();
});

$(window).resize(() => {
  syncInputAndOutputWidths();
});
