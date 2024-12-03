const BASE_API_URL = "/api/v1";

const getTemplateToElement = (tmpl) => {
    const tmplElement = document.createElement("template");
    tmplElement.innerHTML = tmpl.trim();

    return tmplElement.content.firstChild;
};

function showToast(message, type = "success") {
    const toastContainer = document.getElementById("toast-container");

    // Define styles for different types
    const typeStyles = {
        success: "bg-green-500",
        error: "bg-red-500",
        warning: "bg-yellow-500",
    };

    // Create toast element
    const toast = document.createElement("div");
    toast.className = `flex items-center text-white px-4 py-2 rounded shadow-md transition-transform transform translate-x-full opacity-0 ${typeStyles[type] || typeStyles.success}`;
    toast.textContent = message;

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
