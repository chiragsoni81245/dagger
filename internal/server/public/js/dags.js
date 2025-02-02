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

async function submitDag() {
    const DAG_FORM_MODAL = document.getElementById("dag-form-modal");
    let data = {};

    for (let input of DAG_FORM_MODAL.getElementsByTagName("input")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        data[name] = input.value;
    }

    let response = await fetch(`${BASE_API_URL}/dags`, {
        method: "POST",
        headers: { ContentType: "application/json" },
        body: JSON.stringify(data),
    });
    if (response.status != 201) {
        const { error } = await response.json();
        showToast(error, "error");
        return;
    }
    showToast("DAG created successfully");
    closeDagForm();
    renderTable(await getDags());
}

async function renderAddDagForm() {
    const DAG_FORM_MODAL = document.getElementById("dag-form-modal");

    // Reset All Field Values
    for (let input of DAG_FORM_MODAL.getElementsByTagName("input")) {
        input.value = "";
    }

    // Set Title
    document.getElementById("dag-form-title").textContent = "Add Dag";

    DAG_FORM_MODAL.classList.remove("hidden");
}

function closeDagForm() {
    const DAG_FORM_MODAL = document.getElementById("dag-form-modal");
    DAG_FORM_MODAL.classList.add("hidden");
}

function renderTable({ dags, total_dags }) {
    const tableBody = document.getElementById("dags");
    tableBody.innerHTML = "";
    if (dags == null) {
        document.getElementById("result-found").classList.add("hidden");
        document.getElementById("no-result-found").classList.remove("hidden");
        return;
    } else {
        document.getElementById("result-found").classList.remove("hidden");
        document.getElementById("no-result-found").classList.add("hidden");
    }

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
      <td class="py-3 px-6 text-left">${capitalize(row.status)}</td>
      <td class="py-3 px-6 text-left">${row.pending_tasks}</td>
      <td class="py-3 px-6 text-left">${row.running_tasks}</td>
      <td class="py-3 px-6 text-left">${row.completed_tasks}</td>
      <td class="py-3 px-6 text-left">${new Date(row.created_at).toLocaleString()}</td>
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

document
    .getElementById("cancel-dag-form")
    .addEventListener("click", closeDagForm);

document
    .querySelectorAll("button.add-dag")
    .forEach((ele) => ele.addEventListener("click", renderAddDagForm));

document.getElementById("submit-dag-form").addEventListener("click", submitDag);

document
    .getElementById("yaml-config")
    .addEventListener("change", async function () {
        document.getElementById("dag-manual-form").classList.add("hidden");
        document
            .querySelector("#dag-form-modal div.separator")
            .classList.add("hidden");
        document.getElementById("cancel-dag-form").classList.add("hidden");
        document.getElementById("back-dag-form").classList.remove("hidden");

        // Disable Submit button
        document
            .getElementById("submit-dag-form")
            .classList.add("bg-gray-400", "cursor-not-allowed");
        document
            .getElementById("submit-dag-form")
            .classList.remove("bg-blue-600", "hover:bg-blue-700");
        document.getElementById("submit-dag-form").disabled = true;
        // Add Loader
        this.parentElement.querySelector(".spin").classList.remove("hidden");

        // Do validation request to server for this YAML, and to get what inputs to take from user for this DAG definition
        const formData = new FormData();
        formData.append("yaml-config", this.files[0]);

        let response = await fetch(`${BASE_API_URL}/dags/yaml/validate`, {
            method: "POST",
            body: formData,
        });
        const { isValid, error, requiredFiles } = await response.json();
        if (response.status != 200 || !isValid) {
            this.classList.add("border-red-600");
            showToast(error, "error");
        } else {
            console.log({ isValid, requiredFiles });
            showToast("Vaild yaml");

            this.classList.remove("border-red-600");
            // Enable Submit button
            document
                .getElementById("submit-dag-form")
                .classList.remove("bg-gray-400", "cursor-not-allowed");
            document
                .getElementById("submit-dag-form")
                .classList.add("bg-blue-600", "hover:bg-blue-700");
            document.getElementById("submit-dag-form").disabled = false;
        }
        // Remove Loader
        this.parentElement.querySelector(".spin").classList.add("hidden");
    });

document.getElementById("back-dag-form").addEventListener("click", function () {
    const yamlConfig = document.getElementById("yaml-config");
    yamlConfig.value = "";
    document.getElementById("back-dag-form").classList.add("hidden");
    document.getElementById("cancel-dag-form").classList.remove("hidden");
    document.getElementById("dag-manual-form").classList.remove("hidden");
    document
        .querySelector("#dag-form-modal div.separator")
        .classList.remove("hidden");

    // Enable Submit button
    document
        .getElementById("submit-dag-form")
        .classList.remove("bg-gray-400", "cursor-not-allowed");
    document
        .getElementById("submit-dag-form")
        .classList.add("bg-blue-600", "hover:bg-blue-700");
    document.getElementById("submit-dag-form").disabled = false;
    // Remove Loader
    yamlConfig.parentElement.querySelector(".spin").classList.add("hidden");
    yamlConfig.classList.remove("border-red-600");
});

main();
