$(".button, .input, .section").on("touchstart", (event) => {
  if (event.touches.length > 1) {
    event.preventDefault();
  }
});
