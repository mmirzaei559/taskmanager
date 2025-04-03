import { useState, useEffect } from 'react';
import axios from 'axios';
import type { Task } from '@types/task';

function App() {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [newTask, setNewTask] = useState<Omit<Task, 'id' | 'created_at' | 'completed'>>({
        title: '',
        description: '',
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchTasks = async () => {
        setLoading(true);
        try {
            const response = await axios.get('/api/tasks');
            setTasks(response.data);
        } catch (err) {
            setError('Failed to fetch tasks');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTasks();
    }, []);

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
            {/* Your JSX here (same as previous modern version) */}
        </div>
    );
}

export default App;
