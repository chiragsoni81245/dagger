const BASE_API_URL = "/api/v1";
const subscriptions = {};

async function subscribe(eventDescriptor, onmessage) {
    subscriptions[eventDescriptor] = onmessage;
    // Call websocket for this subscription
    window.socket.send(
        JSON.stringify({
            action: "subscribe",
            event: eventDescriptor,
        })
    );
}

async function onmessage(e) {
    let message;
    try {
        message = JSON.parse(e.data);
    } catch {
        console.log("invalid message from socket", message);
        return;
    }
    subscriptions[message["eventDescriptor"]](message);
}

async function main() {
    // Connect to the WebSocket server
    window.socket = new WebSocket(`ws://${window.location.host}/ws`);
    window.socket.onmessage = onmessage;
}

main();

const getTemplateToElement = (tmpl) => {
    const tmplElement = document.createElement("template");
    tmplElement.innerHTML = tmpl.trim();

    return tmplElement.content.firstChild;
};

function showToast(message, type = "success") {
    const toastContainer = document.getElementById("toast-container");

    // Define styles for different types
    const typeStyles = {
        success: "green",
        error: "red",
        warning: "yellow",
    };
    const icon = {
        success: "check",
        error: "times",
        warning: "exclamation-triangle",
    };

    // Create toast element
    const toast = getTemplateToElement(`
        <div class="flex items-center bg-${typeStyles[type]}-100 border border-${typeStyles[type]}-300 rounded-lg shadow-lg p-4">
            <i class="fa fa-${icon[type]} text-${typeStyles[type]}-800" aria-hidden="true"></i>
            <p class="text-sm font-medium text-${typeStyles[type]}-800 ml-2">${message}</p> 
        </div>
    `);

    // Add toast to container
    toastContainer.appendChild(toast);

    // Trigger slide-in animation
    setTimeout(() => {
        toast.classList.remove("translate-x-full", "opacity-0");
        toast.classList.add("translate-x-0", "opacity-100");
    }, 100);

    // Remove toast after 3 seconds
    setTimeout(() => {
        toast.classList.remove("translate-x-0", "opacity-100");
        toast.classList.add("translate-x-full", "opacity-0");
        setTimeout(() => toast.remove(), 300); // Remove element after animation ends
    }, 3000);
}

class Queue {
    constructor() {
        this.items = {}; // Use an object for storage
        this.front = 0; // Tracks the index of the front element
        this.rear = 0; // Tracks the index of the next available position
    }

    // Add an element to the end of the queue
    enqueue(element) {
        this.items[this.rear] = element;
        this.rear++;
    }

    // Remove and return the element at the front of the queue
    dequeue() {
        if (this.isEmpty()) {
            return "Queue is empty";
        }
        const element = this.items[this.front];
        delete this.items[this.front]; // Remove the element
        this.front++; // Move the front pointer
        return element;
    }

    // Return the element at the front without removing it
    peek() {
        if (this.isEmpty()) {
            return "Queue is empty";
        }
        return this.items[this.front];
    }

    // Check if the queue is empty
    isEmpty() {
        return this.front === this.rear;
    }

    // Return the size of the queue
    size() {
        return this.rear - this.front;
    }

    // Clear the queue
    clear() {
        this.items = {};
        this.front = 0;
        this.rear = 0;
    }
}

function capitalize(str) {
    if (!str) return ""; // Handle empty or null strings
    return str.charAt(0).toUpperCase() + str.slice(1);
}

async function getFileContent(url) {
    let res = await fetch(url);
    return await res.text();
}
