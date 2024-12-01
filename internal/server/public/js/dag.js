const DAG_CONTAINER = document.getElementById("dag-container");
const DAG_ID = parseInt(document.getElementById("dag-id").value);
const NAV = document.getElementsByTagName("nav")[0];

const getTemplateToElement = (tmpl) => {
    const tmplElement = document.createElement("template");
    tmplElement.innerHTML = tmpl.trim();

    return tmplElement.content.firstChild;
};

async function submitTask() {
    const EXECUTOR_FORM_MODAL = document.getElementById("task-form-modal");
    const formData = new FormData();

    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("input")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        formData.append(name, input.value);
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("textarea")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        formData.append(name, input.value);
    }
    for (let input of EXECUTOR_FORM_MODAL.getElementsByTagName("select")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        formData.append(name, input.value);
    }

    let response = await fetch(`${BASE_API_URL}/tasks`, {
        method: "POST",
        body: formData,
    });
    if (response.status != 201) {
        // Show error
        return;
    }
    // Show message
    closeTaskForm();
    renderDag(await getDag());
}

async function renderAddTaskForm() {
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    let parent = this.dataset["parent"];
    const executors = await getExecutors();
    if (executors == null) {
        // Show error
        console.log("No executors found, so not opening add task form!");
        return;
    }

    // Reset All Field Values
    for (let input of TASK_FORM_MODAL.getElementsByTagName("input")) {
        input.value = "";
    }
    for (let input of TASK_FORM_MODAL.getElementsByTagName("textarea")) {
        input.value = "";
    }
    for (let input of TASK_FORM_MODAL.getElementsByTagName("select")) {
        input.value = "";
    }

    // Set Title
    document.getElementById("task-form-title").textContent = "Add Task";

    // Set Executors
    const EXECUTOR_CONTAINER = document.getElementById("task-executor-id");
    EXECUTOR_CONTAINER.innerHTML = "";
    EXECUTOR_CONTAINER.appendChild(
        getTemplateToElement(
            `<option value="" disabled selected>Select an executor</option>`
        )
    );
    for (let executor of executors) {
        EXECUTOR_CONTAINER.appendChild(
            getTemplateToElement(
                `<option value="${executor.id}">${executor.name}</option>`
            )
        );
    }

    TASK_FORM_MODAL.classList.remove("hidden");
}

function closeTaskForm() {
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    TASK_FORM_MODAL.classList.add("hidden");
}

async function renderDag(dag) {
    if (dag.tasks == null) {
        let addTaskNode = getTemplateToElement(`
            <div id="add-task" class="w-[200px] task rounded-lg" id="add-task-btn" data-parent="null">
                <p>Add Task</p>
            </div>
        `);
        addTaskNode.addEventListener("click", renderAddTaskForm);
        DAG_CONTAINER.appendChild(addTaskNode);
        return;
    }
}

async function getExecutors() {
    const response = await fetch(`${BASE_API_URL}/executors`);
    if (response.status != 200) {
        // Show error
        return;
    }
    const { executors } = await response.json();
    return executors;
}

async function getDag() {
    const response = await fetch(`${BASE_API_URL}/dags/${DAG_ID}`);
    if (response.status != 200) {
        // Show error
        return;
    }
    const { dag } = await response.json();
    return dag;
}

document
    .getElementById("cancel-task-form")
    .addEventListener("click", closeTaskForm);

document
    .getElementById("submit-task-form")
    .addEventListener("click", submitTask);

async function main() {
    DAG_CONTAINER.style["height"] =
        `${window.innerHeight - NAV.clientHeight}px`;
    renderDag(await getDag());
}

main();
