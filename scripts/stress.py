from locust import HttpUser, task, between

class LoadTest(HttpUser):
    wait_time = between(1, 2)

    @task
    def test_server(self):
        self.client.post("/play", data={"bestOf": 1})


