const DAG_CONTAINER = document.getElementById("dag-container");
const MARGIN_IN_TASKS = 10;
const DAG_ID = parseInt(document.getElementById("dag-id").value);
const NAV = document.getElementsByTagName("nav")[0];

async function submitTask() {
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    const formData = new FormData();

    formData.append("parent_id", TASK_FORM_MODAL.dataset["parent_id"]);
    formData.append("dag_id", DAG_ID);
    for (let input of TASK_FORM_MODAL.getElementsByTagName("input")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        if (input.type == "file") {
            formData.append(name, input.files[0]);
        } else {
            formData.append(name, input.value);
        }
    }
    for (let input of TASK_FORM_MODAL.getElementsByTagName("textarea")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        formData.append(name, input.value);
    }
    for (let input of TASK_FORM_MODAL.getElementsByTagName("select")) {
        if (!input.reportValidity()) return;
        let name = input.getAttribute("name");
        formData.append(name, input.value);
    }

    let response = await fetch(`${BASE_API_URL}/tasks`, {
        method: "POST",
        body: formData,
    });
    if (response.status != 201) {
        const { error } = await response.json();
        showToast(error, "error");
        return;
    }
    showToast("Task created successfully");
    closeTaskForm();
    renderDag(await getDag());
}

async function renderAddTaskForm() {
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    let parent_id = this.dataset["parent_id"];
    TASK_FORM_MODAL.dataset["parent_id"] = parent_id;
    const executors = await getExecutors();
    if (executors == null) {
        showToast("No executors found, so not opening add task form!", "error");
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
    if (!dag) return;
    if (dag.tasks == null) {
        let addTaskNode = getTemplateToElement(`
            <div id="add-task" class="w-[200px] task rounded-lg" id="add-task-btn" data-parent_id="null">
                <p>Add Task</p>
            </div>
        `);
        addTaskNode.addEventListener("click", renderAddTaskForm);
        DAG_CONTAINER.appendChild(addTaskNode);
        return;
    }

    let tasks = {};
    let graph = {};
    let root;
    for (let task of dag.tasks) {
        tasks[task.id] = task;
        if (task.parent_id == null) {
            root = task.id;
            graph[task.id] = [];
        } else if (graph.hasOwnProperty(task.parent_id)) {
            graph[task.parent_id].push(task.id);
        } else {
            graph[task.parent_id] = [task.id];
        }
    }
    console.log(graph);

    // Clear DAG Container
    DAG_CONTAINER.innerHTML = "";

    // BFS Rendering
    q = new Queue();
    q.enqueue(root);

    while (q.size()) {
        let current = q.dequeue();

        // Render Task
        const task = tasks[current];
        console.log(task);
        const taskNode = getTemplateToElement(`
            <div id="task-${task.id}" class="task justify-center items-center p-2" style="left: ${MARGIN_IN_TASKS}; top: ${MARGIN_IN_TASKS}">
                <div class="flex flex-row justify-center items-center">
                    <img src="/static/images/${task.type}-icon.png"/>
                    <p>${task.name}</p>
                </div>
                <div class="flex flex-row justify-center items-center">
                    <svg class="animate-spin h-5 w-5 mr-3 ..." viewBox="0 0 24 24"></svg>
                    <button class="delete-task text-red-500 hover:text-red-700 focus:outline-none">
                        <i class="fa fa-trash" aria-hidden="true"></i>
                    </button>
                </div>
            </div>
        `);
        DAG_CONTAINER.appendChild(taskNode);

        for (let child of graph[current]) {
            q.enqueue(child);
        }
    }
}

async function getExecutors() {
    const response = await fetch(`${BASE_API_URL}/executors`);
    if (response.status != 200) {
        if (response.status == 404) {
            showToast("No executor found", "error");
        } else {
            const { error } = await response.json();
            showToast(error, "error");
        }
        return;
    }
    const { executors } = await response.json();
    return executors;
}

async function getDag() {
    const response = await fetch(`${BASE_API_URL}/dags/${DAG_ID}`);
    if (response.status != 200) {
        if (response.status == 404) {
            showToast("No dag found", "error");
        } else {
            const { error } = await response.json();
            showToast(error, "error");
        }
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
