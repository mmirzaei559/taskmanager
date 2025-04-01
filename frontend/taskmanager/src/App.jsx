import React, { useState, useEffect } from 'react';
import axios from 'axios';

function App() {
    const [tasks, setTasks] = useState([]);
    const [newTask, setNewTask] = useState({ title: '', description: '' });
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        fetchTasks();
    }, []);

    const fetchTasks = async () => {
        setLoading(true);
        try {
            const response = await axios.get('http://localhost:8080/api/tasks');
            setTasks(response.data);
        } catch (error) {
            console.error('Error fetching tasks:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setNewTask({ ...newTask, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await axios.post('http://localhost:8080/api/tasks', newTask);
            setNewTask({ title: '', description: '' });
            fetchTasks();
        } catch (error) {
            console.error('Error creating task:', error);
        }
    };

    const toggleTaskStatus = async (id, currentStatus) => {
        try {
            await axios.post('http://localhost:8080/api/tasks/update', {
                id,
                completed: !currentStatus,
            });
            fetchTasks();
        } catch (error) {
            console.error('Error updating task:', error);
        }
    };

    const runBenchmark = async () => {
        try {
            await axios.get('http://localhost:8080/api/benchmark?count=1000');
            alert('Benchmark completed!');
            fetchTasks();
        } catch (error) {
            console.error('Error running benchmark:', error);
        }
    };

    return (
        <div className="App">
            <h1>Task Manager</h1>

            <button onClick={runBenchmark}>Run Benchmark (1000 tasks)</button>

            <h2>Add New Task</h2>
            <form onSubmit={handleSubmit}>
                <input
                    type="text"
                    name="title"
                    placeholder="Title"
                    value={newTask.title}
                    onChange={handleInputChange}
                    required
                />
                <textarea
                    name="description"
                    placeholder="Description"
                    value={newTask.description}
                    onChange={handleInputChange}
                />
                <button type="submit">Add Task</button>
            </form>

            <h2>Tasks</h2>
            {loading ? (
                <p>Loading...</p>
            ) : (
                <ul>
                    {tasks.map((task) => (
                        <li
                            key={task.id}
                            style={{ textDecoration: task.completed ? 'line-through' : 'none' }}>
                            <h3>{task.title}</h3>
                            <p>{task.description}</p>
                            <small>Created: {new Date(task.created_at).toLocaleString()}</small>
                            <button onClick={() => toggleTaskStatus(task.id, task.completed)}>
                                {task.completed ? 'Mark Incomplete' : 'Mark Complete'}
                            </button>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}

export default App;
