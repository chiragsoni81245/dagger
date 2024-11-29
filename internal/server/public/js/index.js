// Optional JS for dynamic content (e.g., if you want to highlight active sections)
const links = document.querySelectorAll("nav a");
links.forEach((link) => {
    link.addEventListener("click", function () {
        links.forEach((l) =>
            l.classList.remove("font-semibold", "text-gray-200")
        );
        link.classList.add("font-semibold", "text-gray-200");
    });
});
