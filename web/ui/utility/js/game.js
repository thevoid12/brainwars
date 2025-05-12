  document.addEventListener("htmx:afterSwap", function(evt) {
    if (evt.target.classList.contains("navbar")) {
      const navLinks = document.querySelectorAll(".navbar a");
      navLinks.forEach(link => {
        link.classList.add("pointer-events-none", "opacity-50");
        link.setAttribute("title", "Disabled during game");
      });
    }
  });
//     // Prevent F5 and Ctrl+R / Cmd+R
//   window.addEventListener('keydown', function (e) {
//     // F5 key or Ctrl/Cmd + R
//     if (e.key === 'F5' || ((e.ctrlKey || e.metaKey) && e.key === 'r')) {
//       e.preventDefault();
//       alert("Page refresh is disabled during the game.");
//     }
//   });

//   // Disable context menu refresh options (not foolproof, but prevents accidental right-click)
//   window.addEventListener('contextmenu', function (e) {
//     e.preventDefault();
//   });

//   // Prevent drag-to-refresh on touch devices
//   document.addEventListener('touchstart', function (e) {
//     if (e.touches.length > 1) {
//       e.preventDefault();
//     }
//   }, { passive: false });

//   // Before unload (if someone tries to close or reload via browser UI)
//   window.addEventListener('beforeunload', function (e) {
//     e.preventDefault();
//     e.returnValue = '';
//   });