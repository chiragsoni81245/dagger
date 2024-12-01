const DAG_CONTAINER = document.getElementById("dag-container");
const DAG_ID = parseInt(document.getElementById("dag-id").value);
const NAV = document.getElementsByTagName("nav")[0];

const getTemplateToElement = (tmpl) => {
    const tmplElement = document.createElement("template");
    tmplElement.innerHTML = tmpl.trim();

    return tmplElement.content.firstChild;
};

async function addTask() {
    let parent = this.dataset["parent"];
    console.log(parent);
}

async function renderDag(dag) {
    console.log(dag);

    if (dag.tasks == null) {
        let addTaskNode = getTemplateToElement(`
            <div class="w-[200px] task" id="add-task-btn" data-parent="null">
                <p>Add Task</p>
            </div>
        `);
        addTaskNode.addEventListener("click", addTask);
        DAG_CONTAINER.appendChild(addTaskNode);
        return;
    }
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

async function main() {
    DAG_CONTAINER.style["height"] =
        `${window.innerHeight - NAV.clientHeight}px`;
    renderDag(await getDag());
}

main();
