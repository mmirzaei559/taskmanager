export interface Task {
    id: number;
    title: string;
    description: string;
    completed: boolean;
    created_at: string;
}

export interface TaskResult {
    task: Task;
    task_id?: number;
    success: boolean;
    error?: string;
}
