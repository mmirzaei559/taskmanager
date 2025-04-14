import React, { useState, useEffect } from 'react';
import axios from 'axios';
import type { Task, TaskResult } from './types/task';
import './App.css';

function App() {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [newTask, setNewTask] = useState<Omit<Task, 'id' | 'created_at' | 'completed'>>({
        title: '',
        description: '',
    });

    const [bulkTasks, setBulkTasks] = useState<string>('');
    const [processing, setProcessing] = useState(false);
    const [results, setResults] = useState<TaskResult[]>([]);

    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

    console.log(API_BASE_URL);

    const fetchTasks = async () => {
        setIsLoading(true);
        try {
            const response = await axios.get(`${API_BASE_URL}/api/tasks`);
            if (!Array.isArray(response.data)) {
                throw new Error('Invalid tasks data format');
            }
            setTasks(response.data);
        } catch (err) {
            setError('Failed to fetch tasks');
            console.error('Fetch error:', err);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchTasks();
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setNewTask((prev) => ({
            ...prev,
            [name]: value,
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!newTask.title.trim()) {
            setError('Title is required');
            return;
        }

        try {
            await axios.post(`${API_BASE_URL}/api/tasks`, newTask);
            setNewTask({ title: '', description: '' });
            await fetchTasks();
        } catch (err) {
            setError('Failed to create task');
        }
    };

    const toggleTaskStatus = async (id: number, currentStatus: boolean) => {
        try {
            await axios.post(`${API_BASE_URL}/api/tasks/update`, { id, completed: !currentStatus });
            await fetchTasks();
        } catch (err) {
            setError('Failed to update task');
        }
    };

    const runBenchmark = async () => {
        try {
            await axios.get(`${API_BASE_URL}/api/benchmark?count=1000`);
            alert('Benchmark completed!');
            await fetchTasks();
        } catch (err) {
            setError('Benchmark failed');
        }
    };

    const handleBulkSubmit = async () => {
        setProcessing(true);
        try {
            const tasks = bulkTasks
                .split('\n')
                .filter((t) => t.trim())
                .map((title) => ({ title, description: '' }));

            const response = await axios.post(`${API_BASE_URL}/api/tasks/bulk`, tasks);
            setResults(response.data);
        } catch (err) {
            setError('Bulk processing failed');
        } finally {
            setProcessing(false);
        }
    };

    return (
        <div className="app-container">
            <div className="container">
                <header className="header">
                    <h1 className="title">Task Manager</h1>
                    <button onClick={runBenchmark} className="benchmark-btn">
                        Run Benchmark (1000 tasks)
                    </button>
                </header>

                <p className="task-meta">
                    Added from: {task.client_ip || 'Unknown'} • Created:{' '}
                    {new Date(task.created_at).toLocaleString()}
                </p>

                <section className="section">
                    <h2 className="section-title">Bulk Add Tasks</h2>
                    <div className="bulk-section">
                        <textarea
                            value={bulkTasks}
                            onChange={(e) => setBulkTasks(e.target.value)}
                            placeholder="Enter one task per line"
                            rows={5}
                            className="input textarea"
                        />
                        <button
                            onClick={handleBulkSubmit}
                            disabled={processing}
                            className={`submit-btn ${processing ? 'disabled' : ''}`}>
                            {processing ? 'Processing...' : 'Add Tasks Concurrently'}
                        </button>

                        {results.length > 0 && (
                            <div className="results">
                                <h4>Results:</h4>
                                <ul>
                                    {results.map((result, i) => (
                                        <li
                                            key={i}
                                            className={result.success ? 'success' : 'error'}>
                                            {result.task.title}:
                                            {result.success ? ' ✔' : ` ✖ (${result.error})`}
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </div>
                </section>

                <section className="section">
                    <h2 className="section-title">Add New Task</h2>
                    <form onSubmit={handleSubmit}>
                        <div className="form-group">
                            <label htmlFor="title" className="label">
                                Title *
                            </label>
                            <input
                                id="title"
                                name="title"
                                type="text"
                                value={newTask.title}
                                onChange={handleInputChange}
                                className="input"
                                required
                            />
                        </div>

                        <div className="form-group">
                            <label htmlFor="description" className="label">
                                Description
                            </label>
                            <textarea
                                id="description"
                                name="description"
                                rows={3}
                                value={newTask.description}
                                onChange={handleInputChange}
                                className="input textarea"
                            />
                        </div>

                        <button type="submit" className="submit-btn">
                            Add Task
                        </button>
                    </form>
                </section>

                {error && (
                    <div className="error-message">
                        <p>{error}</p>
                    </div>
                )}

                <section className="section">
                    <h2 className="section-title">Tasks</h2>

                    {isLoading ? (
                        <div className="loading-spinner">
                            <div className="spinner"></div>
                        </div>
                    ) : tasks.length === 0 ? (
                        <p className="empty-state">No tasks found. Add one above!</p>
                    ) : (
                        <ul className="task-list">
                            {tasks.map((task) => (
                                <li
                                    key={task.id}
                                    className={`task-item ${task.completed ? 'completed' : ''}`}>
                                    <div className="task-content">
                                        <div>
                                            <h3
                                                className={`task-title ${
                                                    task.completed ? 'completed' : ''
                                                }`}>
                                                {task.title}
                                            </h3>
                                            {task.description && (
                                                <p className="task-description">
                                                    {task.description}
                                                </p>
                                            )}
                                            <p className="task-date">
                                                Created:{' '}
                                                {new Date(task.created_at).toLocaleString()}
                                            </p>
                                        </div>
                                        <button
                                            onClick={() =>
                                                toggleTaskStatus(task.id, task.completed)
                                            }
                                            className={`status-btn ${
                                                task.completed ? 'completed' : 'pending'
                                            }`}>
                                            {task.completed ? 'Mark Incomplete' : 'Mark Complete'}
                                        </button>
                                    </div>
                                </li>
                            ))}
                        </ul>
                    )}
                </section>
            </div>
        </div>
    );
}

export default App;
