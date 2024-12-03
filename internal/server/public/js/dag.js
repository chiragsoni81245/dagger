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

function getTaskNode(task, level, prevTaskId) {
    const parent = document.getElementById(`task-${task.parent_id}`);
    const prevTask = document.getElementById(`task-${prevTaskId}`);
    let [left, top] = [MARGIN_IN_TASKS, MARGIN_IN_TASKS];
    if (parent) {
        left += parent.offsetWidth + parent.offsetLeft;
        if (prevTask) {
            top += prevTask.offsetHeight + prevTask.offsetTop;
        } else {
            top += parent.offsetHeight + parent.offsetTop;
        }
    }

    return getTemplateToElement(`
        <div 
            id="task-${task.id}" 
            class="task bg-blue-500 text-white justify-center" 
            data-level="${level}" 
            style="left: ${left}px; top: ${top}px"
        >
            <div class="flex flex-row justify-start items-center">
                <img src="/static/images/${task.type}-icon.png"/>
                <p class="">${task.name}</p>
            </div>
            <div class="flex flex-row justify-between items-end px-1">
                ${
                    task.status == "processing"
                        ? `<div class="flex items-center justify-center">
                        <div class="w-4 h-4 border-2 border-grey-500 border-t-transparent rounded-full animate-spin"></div>
                    </div>`
                        : `<i class="fa fa-play-circle-o" aria-hidden="true"></i>`
                }
                <div class="flex items-end mx-1">
                    <button class="delete-task flex text-red-500 text-base hover:text-red-700 focus:outline-none">
                        <i class="fa fa-trash" aria-hidden="true"></i>
                    </button>
                    <button class="add-task flex ml-4 text-white text-base hover:text-gray-300 focus:outline-none" data-parent_id="${task.id}">
                        <i class="fa fa-plus" aria-hidden="true"></i>
                    </button>
                </div>
            </div>
        </div>
    `);
}

function drawArrowBetweenTwoTasks(parentTaskId, childTaskId) {
    const parent = document.getElementById(`task-${parentTaskId}`);
    const child = document.getElementById(`task-${childTaskId}`);
    console.log({ parent, child });
    if (!parent || !child) return;
    let startPoint = {
        left: parent.offsetLeft + parent.offsetWidth / 2,
        top: parent.offsetTop + parent.offsetHeight,
    };
    let endPoint = {
        left: child.offsetLeft,
        top: child.offsetTop + child.offsetHeight / 2,
    };
    const arrowHeadWidth = 12;
    const arrowHeadHeight = 6;
    const arrow = getTemplateToElement(`
        <svg class="arrow" width="${DAG_CONTAINER.scrollWidth}" height="${DAG_CONTAINER.scrollHeight}">
            <!-- First line -->
            <line x1="${startPoint.left}" y1="${startPoint.top}" x2="${startPoint.left}" y2="${endPoint.top}" />
            <!-- Second line -->
            <line x1="${startPoint.left}" y1="${endPoint.top}" x2="${endPoint.left}" y2="${endPoint.top}" />
            <!-- Arrowhead -->
            <polygon points="${endPoint.left},${endPoint.top} ${endPoint.left - arrowHeadWidth},${endPoint.top - arrowHeadHeight} ${endPoint.left - arrowHeadWidth},${endPoint.top + arrowHeadHeight}" />
        </svg>
    `);

    DAG_CONTAINER.appendChild(arrow);
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
            graph[task.id] = graph[task.id] || [];
        } else if (graph.hasOwnProperty(task.parent_id)) {
            graph[task.parent_id].push(task.id);
        } else {
            graph[task.parent_id] = [task.id];
        }
    }

    // Clear DAG Container
    DAG_CONTAINER.innerHTML = "";

    // BFS Rendering
    q = [];
    q.push({ node: root, level: 1 });
    let prevTaskId;

    while (q.length > 0) {
        let { node: current, level } = q.pop();

        // Render Task
        const task = tasks[current];
        const taskNode = getTaskNode(task, level, prevTaskId);
        DAG_CONTAINER.appendChild(taskNode);
        document
            .querySelector(`div#task-${task.id} button.add-task`)
            .addEventListener("click", renderAddTaskForm);

        prevTaskId = task.id;

        drawArrowBetweenTwoTasks(task.parent_id, task.id);

        if (!graph[current]) continue;
        for (let i = 0; i < graph[current].length; i++) {
            const child = graph[current][i];
            q.push({ node: child, level: level + 1 });
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
