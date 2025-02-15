const DAG_CONTAINER = document.getElementById("dag-container");
const DAG_HEADER = document.getElementById("dag-header");
const MARGIN_IN_TASKS = 20;
let DAG;
const DAG_ID = parseInt(document.getElementById("dag-id").value);
const NAV = document.getElementsByTagName("nav")[0];

// ---------------------------------- Task Actions ---------------------------------
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

async function deleteTask() {
    const taskId = parseInt(this.dataset["taskId"]);
    const taskName = this.dataset["taskName"];
    // Take confirmation
    if (!confirm(`Are you sure you want to delete '${taskName}' task`)) return;
    const response = await fetch(`${BASE_API_URL}/tasks/${taskId}`, {
        method: "DELETE",
    });
    if (response.status != 200) {
        const { error } = await response.json();
        showToast(error, "error");
        return;
    }
    showToast("Task deleted successfully");
    renderDag(await getDag());
}

// ----------------------------------- Task Form -----------------------------------
async function renderTaskForm(e) {
    e.stopPropagation();
    const action = this.classList.contains("edit-task") ? "Edit" : "Add";
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    let parent_id = this.dataset["parent_id"];
    TASK_FORM_MODAL.dataset["parent_id"] = parent_id;
    const executors = await getExecutors();
    if (executors == null) {
        showToast("No executors found, so not opening add task form!", "error");
        return;
    }

    // Set Title
    document.getElementById("task-form-title").textContent = `${action} Task`;

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

    if (action == "Add") {
        // Reset All Field Values
        for (let input of TASK_FORM_MODAL.getElementsByTagName("input")) {
            if (input.id == "dockerfile-path") {
                input.value = "Dockerfile";
            } else {
                input.value = "";
            }
        }
        for (let input of TASK_FORM_MODAL.getElementsByTagName("textarea")) {
            input.value = "";
        }
        for (let input of TASK_FORM_MODAL.getElementsByTagName("select")) {
            input.value = "";
        }
    } else {
    }

    TASK_FORM_MODAL.classList.remove("hidden");
}

function closeTaskForm() {
    const TASK_FORM_MODAL = document.getElementById("task-form-modal");
    TASK_FORM_MODAL.classList.add("hidden");
}

// -------------------------------- Task Logs Modal --------------------------------
async function showTaskDetails(e) {
    e.stopPropagation();
    if (
        !this.querySelector(".status-completed") &&
        !this.querySelector(".status-error")
    ) {
        return;
    }
    const TASK_DETAILS_MODAL = document.getElementById("task-details-modal");
    const MENU_BLOCK = TASK_DETAILS_MODAL.querySelector(".menu-actions");
    TASK_DETAILS_MODAL.classList.remove("hidden");

    TASK_DETAILS_MODAL.querySelector("h2.title").textContent =
        `${this.querySelector("p.name").textContent} Logs`;

    const detailsBlock = TASK_DETAILS_MODAL.querySelector("div.details");
    const taskId = this.id.split("-")[1];
    const logs = await getTaskLogs(taskId);

    detailsBlock.innerHTML = "";
    if (logs.length == 0) {
        detailsBlock.appendChild(getTemplateToElement(`<p>No logs found</p>`));
        return;
    }

    MENU_BLOCK.querySelector(".expand")?.remove();
    MENU_BLOCK.insertBefore(
        getTemplateToElement(
            `<i class="expand fa fa-window-maximize cursor-pointer text-gray-400 mx-2 rounded" aria-hidden="true"></i>`
        ),
        document.getElementById("task-details-close")
    );
    detailsBlock.appendChild(
        getTemplateToElement(`
    <div class="tab-links border-b border-gray-200">
        <nav class="flex space-x-4">
            ${Array.from(logs)
                .map(
                    (log, index) => `
                <button 
                    class="tab-link ${index == 0 ? "border-blue-500 text-blue-500" : "text-gray-600"} py-2 px-4 border-b-2 border-transparent hover:border-blue-500 focus:outline-none focus:border-blue-500" 
                    data-tab="task-log-${log.name}"
                >
                    ${log.name}
                </button>
            `
                )
                .join("")}
        </nav>
    </div>`)
    );

    for (let i = 0; i < logs.length; i++) {
        detailsBlock.appendChild(
            getTemplateToElement(`
                <div id="task-log-${logs[i].name}" class="tab-content ${i == 0 ? "" : "hidden"}">
                    <pre class="bg-gray-100 p-4 overflow-auto rounded h-[400px]">${await getFileContent(logs[i].url)}</pre>
                </div>
            `)
        );
    }

    const tabLinks = document.querySelectorAll(".tab-link");
    const tabContents = document.querySelectorAll(".tab-content");

    MENU_BLOCK.querySelector(".expand").addEventListener("click", function () {
        if (this.classList.contains("fa-window-maximize")) {
            this.classList.remove("fa-window-maximize");
            this.classList.add("fa-window-minimize");
            TASK_DETAILS_MODAL.children[0].classList.add("w-[100%]");
            TASK_DETAILS_MODAL.children[0].classList.remove("h-[400px]");
            TASK_DETAILS_MODAL.children[0].classList.add(
                `h-[${window.innerHeight}px]`
            );
            for (let pre of detailsBlock.getElementsByTagName("pre")) {
                pre.classList.add(
                    `h-[${window.innerHeight - TASK_DETAILS_MODAL.querySelector("div.header").clientHeight - detailsBlock.querySelector("div.tab-links").clientHeight}px]`
                );
            }
        } else {
            this.classList.remove("fa-window-minimize");
            this.classList.add("fa-window-maximize");
            TASK_DETAILS_MODAL.children[0].classList.remove("w-[100%]");
            TASK_DETAILS_MODAL.children[0].classList.add("h-[400px]");
            TASK_DETAILS_MODAL.children[0].classList.remove(
                `h-[${window.innerHeight}px]`
            );
            for (let pre of detailsBlock.getElementsByTagName("pre")) {
                pre.classList.remove(
                    `h-[${window.innerHeight - TASK_DETAILS_MODAL.querySelector("div.header").clientHeight - detailsBlock.querySelector("div.tab-links").clientHeight - 40}px]`
                );
            }
        }
    });

    tabLinks.forEach((link) => {
        link.addEventListener("click", () => {
            // Hide all tab contents
            tabContents.forEach((content) => content.classList.add("hidden"));

            // Remove active state from all tabs
            tabLinks.forEach((tab) =>
                tab.classList.remove(
                    "border-blue-500",
                    "text-blue-500",
                    "text-gray-600"
                )
            );

            // Show the clicked tab content
            const tabId = link.getAttribute("data-tab");
            document.getElementById(tabId).classList.remove("hidden");

            // Set active state on the clicked tab
            link.classList.add("border-blue-500", "text-blue-500");
        });
    });

    detailsBlock.querySelector(".tab-link").focus();
}

function closeTaskDetails() {
    const TASK_DETAILS_MODAL = document.getElementById("task-details-modal");
    const MENU_BLOCK = TASK_DETAILS_MODAL.querySelector(".menu-actions");
    const expandButton = MENU_BLOCK.querySelector(".expand");

    TASK_DETAILS_MODAL.classList.add("hidden");
    expandButton.classList.remove("fa-compress");
    expandButton.classList.add("fa-arrows-alt");
    TASK_DETAILS_MODAL.children[0].classList.remove("w-[100%]");
    TASK_DETAILS_MODAL.children[0].classList.remove(
        `h-[${window.innerHeight}px]`
    );
    for (let pre of detailsBlock.getElementsByTagName("pre")) {
        pre.classList.remove(
            `h-[${window.innerHeight - TASK_DETAILS_MODAL.querySelector("div.header").clientHeight - detailsBlock.querySelector("div.tab-links").clientHeight - 40}px]`
        );
    }
}

// ---------------------------------- Dag Render -----------------------------------
async function renderDag(dag) {
    if (!dag) return;
    // Set Dag metadata
    DAG = dag;
    document.getElementById("dag-name").textContent = dag.name;

    // Clear DAG Container
    DAG_CONTAINER.innerHTML = "";
    if (dag.status == "created") {
        if (dag.tasks != null) {
            document.getElementById("run-dag").classList.remove("hidden");
        }
        document.getElementById("delete-dag").classList.remove("hidden");
    } else if (dag.status == "running") {
        document.getElementById("run-dag").classList.add("hidden");
        document.getElementById("delete-dag").classList.add("hidden");
    } else if (dag.status == "completed") {
        document.getElementById("run-dag").classList.add("hidden");
        document.getElementById("delete-dag").classList.remove("hidden");
    }

    if (dag.tasks == null) {
        let addTaskNode = getTemplateToElement(`
            <div id="add-task" class="task p-2 flex flex-row bg-[#4DA8DA] text-white justify-center items-center m-4" id="add-task-btn" data-parent_id="null">
                <i class="fa fa-plus mx-2" aria-hidden="true"></i>
                <p class="mx-2">Add Task</p>
            </div>
        `);
        addTaskNode.addEventListener("click", renderTaskForm);
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
            .querySelector(`div#task-${task.id}`)
            ?.addEventListener("click", showTaskDetails);
        document
            .querySelector(`div#task-${task.id} button.add-task`)
            ?.addEventListener("click", renderTaskForm);
        document
            .querySelector(`div#task-${task.id} button.edit-task`)
            ?.addEventListener("click", renderTaskForm);
        document
            .querySelector(`div#task-${task.id} button.delete-task`)
            ?.addEventListener("click", deleteTask);

        prevTaskId = task.id;

        drawArrowBetweenTwoTasks(task.parent_id, task.id);

        if (!graph[current]) continue;
        for (let i = 0; i < graph[current].length; i++) {
            const child = graph[current][i];
            q.push({ node: child, level: level + 1 });
        }
    }
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
        class="task bg-[#4DA8DA] text-white justify-center ${["completed", "error"].indexOf(task.status) != -1 ? "cursor-pointer" : ""}" 
            data-level="${level}" 
            style="left: ${left}px; top: ${top}px"
        >
            <div class="flex flex-row justify-start items-center mb-2">
                <img src="/static/images/${task.type}-icon.png"/>
                <p class="name ml-auto">${task.name}</p>
            </div>
            <div class="flex flex-row justify-between items-end">
                <div id="task-${task.id}-status">
                    ${getTaskStatusElementString(task.status)}
                </div>
                <div id="task-${task.id}-actions" class="task-action-buttons flex items-end ${DAG.status != "created" ? "hidden" : ""}">
                    <button class="edit-task flex text-white text-base hover:text-grey-300 focus:outline-none hidden" data-task-id="${task.id}" data-task-name="${task.name}">
                        <i class="fa fa-edit" aria-hidden="true"></i>
                    </button>
                    <button class="delete-task flex ml-3 text-[#FF6B6B] text-base hover:text-red-700 focus:outline-none" data-task-id="${task.id}" data-task-name="${task.name}">
                        <i class="fa fa-trash" aria-hidden="true"></i>
                    </button>
                    <button class="add-task flex ml-3 text-white text-base hover:text-gray-300 focus:outline-none" data-parent_id="${task.id}">
                        <i class="fa fa-plus" aria-hidden="true"></i>
                    </button>
                </div>
            </div>
        </div>
    `);
}

function getTaskStatusElementString(status) {
    return {
        created: `
            <div class="flex flex-row text-balck-800">
            <i class="fa fa-list" aria-hidden="true"></i>
            </div>
            `,
        running: `
            <div class="status-running flex flex-row text-white">
            <i class="fa fa-spinner spin" aria-hidden="true"></i>
            </div>
            `,
        error: `
            <div class="status-error flex flex-row text-red-300">
            <i class="fa fa-exclamation-circle" aria-hidden="true"></i>
            </div>
            `,
        completed: `
            <div class="status-completed flex flex-row text-green-300">
            <i class="fa fa-check-circle-o" aria-hidden="true"></i>
            </div>
            `,
    }[status];
}

function drawArrowBetweenTwoTasks(parentTaskId, childTaskId) {
    const parent = document.getElementById(`task-${parentTaskId}`);
    const child = document.getElementById(`task-${childTaskId}`);
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

// --------------------------------- API Calls -------------------------------------
async function getExecutors() {
    const response = await fetch(`${BASE_API_URL}/executors`);
    if (response.status != 200) {
        const { error } = await response.json();
        showToast(error, "error");
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

async function getTask(taskId) {
    const response = await fetch(`${BASE_API_URL}/tasks/${taskId}`);
    if (response.status != 200) {
        if (response.status == 404) {
            showToast("No task found", "error");
        } else {
            const { error } = await response.json();
            showToast(error, "error");
        }
        return;
    }
    const { task } = await response.json();
    return task;
}

async function getTaskLogs(taskId) {
    const response = await fetch(`${BASE_API_URL}/tasks/${taskId}/logs`);
    if (response.status != 200) {
        if (response.status == 404) {
            showToast("No logs found for task", "error");
        } else {
            const { error } = await response.json();
            showToast(error, "error");
        }
        return [];
    }
    const { logs } = await response.json();
    return logs;
}

async function deleteDag(e) {
    e.stopPropagation();
    // Take confirmation
    if (!confirm(`Are you sure you want to delete '${DAG.name}' dag`)) return;
    const response = await fetch(`${BASE_API_URL}/dags/${DAG_ID}`, {
        method: "DELETE",
    });
    if (response.status != 200) {
        const { error } = await response.json();
        showToast(error, "error");
        return;
    }
    window.location.href = "/dags";
}

async function runDag() {
    // Take confirmation
    if (!confirm(`Are you sure you want to start '${DAG.name}' dag`)) return;
    const subscriptionEvent = `dag:${DAG_ID}`;
    subscribe(subscriptionEvent, handleDagSubscriptionMessage);

    const response = await fetch(`${BASE_API_URL}/dags/${DAG_ID}/run`, {
        method: "POST",
    });
    if (response.status != 200) {
        const { error } = await response.json();
        showToast(error, "error");
        return;
    }
    document.getElementById("run-dag").classList.add("hidden");
    document.getElementById("delete-dag").classList.add("hidden");
    for (let node of document.getElementsByClassName("task-action-buttons")) {
        node.classList.add("hidden");
    }
    showToast("Started running");
}

// -------------------------- Websocket Event Handler -----------------------------
async function handleDagSubscriptionMessage(message) {
    if (message.resource == "task") {
        if (message.field == "status") {
            let taskStatusBlock = document.getElementById(
                `task-${message.id}-status`
            );
            taskStatusBlock.innerHTML = getTaskStatusElementString(
                message.newValue
            ).trim();
            if (["completed", "error"].indexOf(message.newValue) != -1) {
                document
                    .getElementById(`task-${message.id}`)
                    .classList.add("cursor-pointer");
            }
        }
    } else if (message.resource == "dag") {
        if (message.field == "status") {
            if (message.newValue == "completed") {
                document.getElementById("run-dag").classList.add("hidden");
                document
                    .getElementById("delete-dag")
                    .classList.remove("hidden");
            }
        }
    }
}

// ------------------------------ Event Listeners ---------------------------------
document
    .getElementById("cancel-task-form")
    .addEventListener("click", closeTaskForm);

document
    .getElementById("submit-task-form")
    .addEventListener("click", submitTask);

document
    .getElementById("task-details-close")
    .addEventListener("click", closeTaskDetails);

document.getElementById("delete-dag").addEventListener("click", deleteDag);
document.getElementById("run-dag").addEventListener("click", runDag);
document.getElementById("task-type").addEventListener("change", (e) => {
    let definationBlocks = document.querySelectorAll(
        `#taks-definition-block div.dynamic-fields > div`
    );
    for (let df of definationBlocks) {
        if (df.classList.contains(`type-${e.target.value}`)) {
            df.classList.remove("hidden");
        } else {
            df.classList.add("hidden");
        }
    }
});

async function main() {
    const dag = await getDag();

    // Subscribe to dag events if dag is in running state
    if (dag.status == "running") {
        subscribe(`dag:${DAG_ID}`, handleDagSubscriptionMessage);
    }
    await renderDag(dag);
    DAG_CONTAINER.style["height"] =
        `${window.innerHeight - NAV.clientHeight - DAG_HEADER.clientHeight}px`;
}

main();
main();
