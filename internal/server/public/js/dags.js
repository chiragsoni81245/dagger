let currentPage = 1;
const pageSize = 10;

async function getDags() {
    let response = await fetch(
        `${BASE_API_URL}/dags?page=${currentPage}&page_size=${pageSize}`
    );
    return await response.json();
}

async function main() {
    renderTable(await getDags());
}

function renderTable({ dags, total_dags }) {
    const tableBody = document.getElementById("dags");
    tableBody.innerHTML = "";

    dags.forEach((row) => {
        const tr = document.createElement("tr");
        tr.id = `dag-${row.id}`;
        tr.classList.add(
            "border-b",
            "border-gray-200",
            "hover:bg-gray-200",
            "hover:cursor-pointer",
            "dag"
        );
        tr.innerHTML = `
      <td class="py-3 px-6 text-left">${row.id}</td>
      <td class="py-3 px-6 text-left">${row.name}</td>
      <td class="py-3 px-6 text-left">${row.status}</td>
      <td class="py-3 px-6 text-left">${row.pending_tasks}</td>
      <td class="py-3 px-6 text-left">${row.completed_tasks}</td>
      <td class="py-3 px-6 text-left">${row.processing_tasks}</td>
      <td class="py-3 px-6 text-left">${row.created_at}</td>
    `;
        tr.addEventListener("click", () => {
            window.location.href = `/dags/${row.id}`;
        });
        tableBody.appendChild(tr);
    });

    const paginationInfo = document.getElementById("pagination-info");
    paginationInfo.textContent = `Page ${currentPage} of ${Math.ceil(total_dags / pageSize)}`;
    const nextBtn = document.getElementById("next-btn");
    const prevBtn = document.getElementById("prev-btn");
    if (Math.ceil(total_dags / pageSize) == 1) {
        nextBtn.style["display"] = "none";
        prevBtn.style["display"] = "none";
    } else if (currentPage == Math.ceil(total_dags / pageSize)) {
        nextBtn.style["display"] = "inline";
        prevBtn.style["display"] = "none";
    } else if (currentPage <= 1) {
        nextBtn.style["display"] = "none";
        prevBtn.style["display"] = "inline";
    } else {
        nextBtn.style["display"] = "inline";
        prevBtn.style["display"] = "inline";
    }
}

document.getElementById("prev-btn").addEventListener("click", async () => {
    if (currentPage > 1) {
        currentPage--;
        renderTable(await getDags());
    }
});

document.getElementById("next-btn").addEventListener("click", async () => {
    if (currentPage < Math.ceil(data.length / pageSize)) {
        currentPage++;
        renderTable(await getDags());
    }
});

main();
