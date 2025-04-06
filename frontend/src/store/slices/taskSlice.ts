import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import axios from 'axios';
import type { Task } from '@/types/task';

interface TaskState {
    tasks: Task[];
    loading: boolean;
    error: string | null;
}

const initialState: TaskState = {
    tasks: [],
    loading: false,
    error: null,
};

// Async thunks for API calls
export const fetchTasks = createAsyncThunk('tasks/fetchAll', async (_, { rejectWithValue }) => {
    try {
        const response = await axios.get('/api/tasks');
        return response.data;
    } catch (err) {
        return rejectWithValue('Failed to fetch tasks');
    }
});

export const createTask = createAsyncThunk(
    'tasks/create',
    async (task: { title: string; description: string }, { rejectWithValue }) => {
        try {
            const response = await axios.post('/api/tasks', task);
            return response.data;
        } catch (err) {
            return rejectWithValue('Failed to create task');
        }
    }
);

export const updateTaskStatus = createAsyncThunk(
    'tasks/updateStatus',
    async ({ id, completed }: { id: number; completed: boolean }, { rejectWithValue }) => {
        try {
            await axios.post('/api/tasks/update', { id, completed });
            return { id, completed };
        } catch (err) {
            return rejectWithValue('Failed to update task');
        }
    }
);

const taskSlice = createSlice({
    name: 'tasks',
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder
            // Fetch Tasks
            .addCase(fetchTasks.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchTasks.fulfilled, (state, action: PayloadAction<Task[]>) => {
                state.loading = false;
                state.tasks = action.payload;
            })
            .addCase(fetchTasks.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            })

            // Create Task
            .addCase(createTask.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(createTask.fulfilled, (state, action: PayloadAction<Task>) => {
                state.loading = false;
                state.tasks.push(action.payload);
            })
            .addCase(createTask.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            })

            // Update Task Status
            .addCase(updateTaskStatus.fulfilled, (state, action) => {
                const { id, completed } = action.payload;
                const task = state.tasks.find((task) => task.id === id);
                if (task) {
                    task.completed = completed;
                }
            });
    },
});

export default taskSlice.reducer;
