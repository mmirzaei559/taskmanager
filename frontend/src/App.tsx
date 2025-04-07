import { useState, useEffect } from 'react';
import axios from 'axios';
import type { Task } from './types/task';

function App() {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [newTask, setNewTask] = useState<Omit<Task, 'id' | 'created_at' | 'completed'>>({
        title: '',
        description: '',
    });

    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(false);

    const fetchTasks = async () => {
        setIsLoading(true);
        try {
            const response = await axios.get('/api/tasks');
            setTasks(response.data);
        } catch (err) {
            setError('Failed to fetch tasks');
            console.error(err);
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
            await axios.post('/api/tasks', newTask);
            setNewTask({ title: '', description: '' });
            await fetchTasks();
        } catch (err) {
            setError('Failed to create task');
        }
    };

    const toggleTaskStatus = async (id: number, currentStatus: boolean) => {
        try {
            await axios.post('/api/tasks/update', { id, completed: !currentStatus });
            await fetchTasks();
        } catch (err) {
            setError('Failed to update task');
        }
    };

    const runBenchmark = async () => {
        try {
            await axios.get('/api/benchmark?count=1000');
            alert('Benchmark completed!');
            await fetchTasks();
        } catch (err) {
            setError('Benchmark failed');
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 p-6">
            <div className="max-w-3xl mx-auto">
                <header className="mb-8">
                    <h1 className="text-3xl font-bold text-gray-800">Task Manager</h1>
                    <button
                        onClick={runBenchmark}
                        className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
                        Run Benchmark (1000 tasks)
                    </button>
                </header>

                <section className="mb-8 p-6 bg-white rounded-lg shadow">
                    <h2 className="text-xl font-semibold mb-4 text-gray-700">Add New Task</h2>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div>
                            <label
                                htmlFor="title"
                                className="block text-sm font-medium text-gray-700 mb-1">
                                Title *
                            </label>
                            <input
                                id="title"
                                name="title"
                                type="text"
                                value={newTask.title}
                                onChange={handleInputChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                                required
                            />
                        </div>

                        <div>
                            <label
                                htmlFor="description"
                                className="block text-sm font-medium text-gray-700 mb-1">
                                Description
                            </label>
                            <textarea
                                id="description"
                                name="description"
                                rows={3}
                                value={newTask.description}
                                onChange={handleInputChange}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                        </div>

                        <button
                            type="submit"
                            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition-colors">
                            Add Task
                        </button>
                    </form>
                </section>

                {error && (
                    <div className="mb-4 p-4 bg-red-100 border-l-4 border-red-500 text-red-700">
                        <p>{error}</p>
                    </div>
                )}

                <section>
                    <h2 className="text-xl font-semibold mb-4 text-gray-700">Tasks</h2>

                    {isLoading ? (
                        <div className="flex justify-center items-center py-8">
                            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
                        </div>
                    ) : tasks.length === 0 ? (
                        <p className="text-gray-500 italic">No tasks found. Add one above!</p>
                    ) : (
                        <ul className="space-y-4">
                            {tasks.map((task) => (
                                <li
                                    key={task.id}
                                    className={`p-4 bg-white rounded-lg shadow ${
                                        task.completed ? 'opacity-75' : ''
                                    }`}>
                                    <div className="flex justify-between items-start">
                                        <div>
                                            <h3
                                                className={`font-medium ${
                                                    task.completed
                                                        ? 'line-through text-gray-500'
                                                        : 'text-gray-800'
                                                }`}>
                                                {task.title}
                                            </h3>
                                            {task.description && (
                                                <p className="mt-1 text-gray-600">
                                                    {task.description}
                                                </p>
                                            )}
                                            <p className="mt-2 text-sm text-gray-400">
                                                Created:{' '}
                                                {new Date(task.created_at).toLocaleString()}
                                            </p>
                                        </div>
                                        <button
                                            onClick={() =>
                                                toggleTaskStatus(task.id, task.completed)
                                            }
                                            className={`px-3 py-1 rounded text-sm ${
                                                task.completed
                                                    ? 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                                                    : 'bg-green-100 text-green-800 hover:bg-green-200'
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
