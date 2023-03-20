import requests

# Define the TaskRequest structure

# input_data = {
#     "start": "2023-03-05T14:00:00Z",
#     "duration": 10,
#     "task": "xxx",
#     "project": "xx"
# }

start = input_data["start"]
task = input_data["task"]
duration = int(input_data["duration"])
project = input_data["project"]

task_request = {
    "start_time": start,
    "duration": duration,
    "task": task,
    "project": project,
}

# Send the POST request with the TaskRequest JSON payload
response = requests.post("https://api.gohi789.top/tasks", json=task_request)

# Check the response status code
if response.status_code == 200:
    print("Task created successfully!")
else:
    print(f"Error creating task: {response.text}")

return {"result": response.text}
