# Pre-configured Jenkins for IaCConsole-CLI IaCConsole-DB/API Demo

This project provides a way to deploy a pre-configured Jenkins instance to any Kubernetes cluster (especially local ones like Kind or Minikube). It's designed to demonstrate an end-to-end workflow using IaCConsole-CLI and IaCConsole-DB with a demo OpenTofu unit.

## Steps to setup

1.  Execute the deployment script:
    ```bash
    bash deploy.sh
    ```

2.  Get the password for the Jenkins admin user:
    ```bash
    kubectl exec --namespace jenkins -it svc/jenkins-dev -c jenkins -- /bin/cat /run/secrets/additional/chart-admin-password && echo
    ```

3.  Start port-forwarding to access the Jenkins UI:
    ```bash
    kubectl port-forward --namespace jenkins svc/jenkins-dev 8080:8080
    ```

4.  Navigate your browser to [http://127.0.0.1:8080](http://127.0.0.1:8080).

5.  Log in with the username `admin` and the password obtained in step 2.

6.  Go to the `iacconsole-pipeline` job page: [http://127.0.0.1:8080/job/iacconsole-pipeline/](http://127.0.0.1:8080/job/iacconsole-pipeline/).

7.  On the left menu, click on **Build**. This will download and execute the `Jenkinsfile` from the repository.

8.  Build #1 should start. Click on the build link to open the console output: [http://127.0.0.1:8080/job/iacconsole-pipeline/1/console](http://127.0.0.1:8080/job/iacconsole-pipeline/1/console).

9.  Scroll to the end of the console output. You should see a plan output from OpenTofu and a question: "Do you want to apply the changes?".

10. Click **Proceed** and watch as the apply stage is performed.

<img width="1432" height="1188" alt="2025-11-07_20-41" src="https://github.com/user-attachments/assets/2d7ea70e-f123-48d7-96fb-5205b04b4238" />
<img width="2924" height="1306" alt="2025-11-07_20-41_1" src="https://github.com/user-attachments/assets/5a0fda74-b0d4-4370-9627-e60cd41109c9" />
<img width="3258" height="1724" alt="2025-11-07_20-42" src="https://github.com/user-attachments/assets/2079b624-f21b-482d-afdb-64efddeb0d2a" />
<img width="3118" height="1214" alt="2025-11-07_20-42_1" src="https://github.com/user-attachments/assets/554cc744-ca25-4c43-8913-40678874570f" />
