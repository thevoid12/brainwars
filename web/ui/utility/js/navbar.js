  function highlightActiveTab() {
    const currentPath = window.location.pathname;
    let bestMatch = null;
    let bestLength = -1;

    const links = document.querySelectorAll('.nav-link');
    if (!links.length) {
    //  console.log("No nav links found. Script ran before sidebar loaded?");
      return;
    }

    links.forEach(link => {
      const href = link.getAttribute('href');
      if (currentPath.startsWith(href) && href.length > bestLength) {
        bestMatch = link;
        bestLength = href.length;
      }
    });

    if (bestMatch) {
      bestMatch.classList.add('bg-primary-50', 'text-primary-600', 'font-medium');
      bestMatch.classList.remove('text-gray-600');
   //   console.log("Active tab set for:", bestMatch.getAttribute('href'));
    } else {
     // console.log("No matching nav link found for path:", currentPath);
    }
  }

  // Wait until everything is loaded (important if sidebar is HTMX-injected)
  document.addEventListener("DOMContentLoaded", highlightActiveTab);
  document.addEventListener("htmx:afterSwap", highlightActiveTab); // In case sidebar loads via HTMX