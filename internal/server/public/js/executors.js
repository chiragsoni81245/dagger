let currentPage = 1;
const pageSize = 10;

async function getExecutors() {
    let response = await fetch(
        `${BASE_API_URL}/executors?page=${currentPage}&page_size=${pageSize}`
    );
    return await response.json();
}

async function main() {
    renderTable(await getExecutors());
}

async function submitExecutor() {
    const EXECUTOR_FORM_MODAL = document.getElementById("executor-form-modal");
    let data = {};

    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("input")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        data[name] = input.value;
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("textarea")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        data[name] = input.value;
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("select")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        if (name == "config") {
            try {
                JSON.parse(input.value);
            } catch {
                // Show error
                return;
            }
        }
        data[name] = input.value;
    }

    let response = await fetch(`${BASE_API_URL}/executors`, {
        method: "POST",
        headers: { ContentType: "application/json" },
        body: JSON.stringify(data),
    });
    if (response.status != 201) {
        // Show error
        return;
    }
    // Show message
    closeExecutorForm();
    renderTable(await getExecutors());
}

async function renderAddExecutorForm() {
    const EXECUTOR_FORM_MODAL = document.getElementById("executor-form-modal");

    // Reset All Field Values
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("input")) {
        input.value = "";
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("textarea")) {
        input.value = "";
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("select")) {
        input.value = "";
    }

    // Set Title
    document.getElementById("executor-form-title").textContent = "Add Executor";

    EXECUTOR_FORM_MODAL.classList.remove("hidden");
}

function closeExecutorForm() {
    const EXECUTOR_FORM_MODAL = document.getElementById("executor-form-modal");
    EXECUTOR_FORM_MODAL.classList.add("hidden");
}

function renderTable({ executors, total_executors }) {
    const tableBody = document.getElementById("executors");
    tableBody.innerHTML = "";
    if (executors == null) {
        document.getElementById("result-found").classList.add("hidden");
        document.getElementById("no-result-found").classList.remove("hidden");
        return;
    } else {
        document.getElementById("result-found").classList.remove("hidden");
        document.getElementById("no-result-found").classList.add("hidden");
    }

    executors.forEach((row) => {
        const tr = document.createElement("tr");
        tr.id = `executor-${row.id}`;
        tr.classList.add(
            "border-b",
            "border-gray-200",
            "hover:bg-gray-200",
            "hover:cursor-pointer",
            "executor"
        );
        tr.innerHTML = `
      <td class="py-3 px-6 text-left">${row.id}</td>
      <td class="py-3 px-6 text-left">${row.name}</td>
      <td class="py-3 px-6 text-left">${row.type}</td>
      <td class="py-3 px-6 text-left">${row.created_at}</td>
    `;
        tableBody.appendChild(tr);
    });

    const paginationInfo = document.getElementById("pagination-info");
    paginationInfo.textContent = `Page ${currentPage} of ${Math.ceil(total_executors / pageSize)}`;
    const nextBtn = document.getElementById("next-btn");
    const prevBtn = document.getElementById("prev-btn");
    if (Math.ceil(total_executors / pageSize) == 1) {
        nextBtn.style["display"] = "none";
        prevBtn.style["display"] = "none";
    } else if (currentPage == Math.ceil(total_executors / pageSize)) {
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
        renderTable(await getExecutors());
    }
});

document.getElementById("next-btn").addEventListener("click", async () => {
    if (currentPage < Math.ceil(data.length / pageSize)) {
        currentPage++;
        renderTable(await getExecutors());
    }
});

document
    .getElementById("cancel-executor-form")
    .addEventListener("click", closeExecutorForm);

document
    .querySelectorAll("button.add-executor")
    .forEach((ele) => ele.addEventListener("click", renderAddExecutorForm));

document
    .getElementById("submit-executor-form")
    .addEventListener("click", submitExecutor);

main();
