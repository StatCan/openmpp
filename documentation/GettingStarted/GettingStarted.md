
# Getting started with Kubeflow Notebooks and OpenM++.


Getting started with Kubeflow Notebooks and OpenM++.
This document is intended to act as an introduction to The AAWs Kubeflow Notebook servers and their use for the OpenM++ project.


## Starting

To access the AAW Kubeflow portal, navigate to the following website.

[https://kubeflow.aaw.cloud.statcan.ca/](https://kubeflow.aaw.cloud.statcan.ca/) 

This will redirect you to a Microsoft log-in page.

![Login01](Login01.png)

Select the account you wish to use and proceed with the authentication.

After your credentials are authenticated, you will be redirected to the AAW Kubeflow management panel. 

The Kubeflow management panel

![Kube Flow Management Panel](KFMP01.png)

## Create a notebook.

Click on the Create a New Notebook server button.

![Create Notebook01](CreateNB01.png)

This brings up the new Notebook screen.

To create a new Notebook, three things must be set:
- Ensure the correct Namespace is selected.
- A unique name in the name field. A timestamped name is automatically generated in the Name field (timestamped) when you click in it,
- then click on the Notebook type you want.
  - For OpenM++, select the JupyterLab option.
- If you are working with Protected B Data, select the **Run in Protected B notebook** checkbox

Scroll down to see the **Launch Button.**  The Launch button will only be active after the above options are selected.
 
![Notebook screen](NewNBScreen02.png)

Press the Launch button to launch you new Notebook.

[For more information about Statcan AAW Kubeflow, Click here.](https://statcan.github.io/aaw/en/1-Experiments/Kubeflow/)

### Existing Notebooks

If you have previously created a notebook, you can reuse it by clicking on the Notebooks Tab.

![Notebook screen](NewNBScreen03.png)

This will bring up a window with all your existing Notebooks.  

![Notebook screen](NewNBScreen04.png)

To start an existing Notebook, select it and press the CONNECT button.

If the Connect button is disabled, click on the triangle (Start image) button to start the image, and then connect when it becomes avaliable 

## Your Kubeflow notebook.

![Kube Flow screen](KFNotebook01.png)

To start the OpemM++ UI, simply click on the OpenM++ icon on the Notebooks page.

![Kube Flow screen](KFNotebook09.png)

This will open a new window with the OpenM++ UI running.

## OpenM++ UI

![OpenM UI screen](OpenMUI01.png)


Click on the Ellipses symbol on the upper Left corner to change language.

![OpenM UI screen](OpenMUI02.png)



Click on the Hamburger Menu on the top right to open the sidebar.

![OpenM UI screen](OpenMUI03.png)



Click on the Model you want (Left Panel) to select it, In this case the IDMM Model.  This brings up the Model Run Panel and activates the Input Scenarios and Run the Model tabs on the Right Panel.

![OpenM UI screen](OpenMUI04.png)

The horizontal tabs are also active (but greyed out) at this time.

![OpenM UI screen](OpenMUI05.png)

To Run a Model, first the Model name must be entered.  Simply clicking in the Model Name box will generate a uniquely timestamped Model name for the run.

![OpenM UI screen](OpenMUI06.png)

You can then click the Run the Model to run the job.

Note. add section on selecting template!

This brings up the Model Run Results Panel which shows the results of the run.

![OpenM UI screen](OpenMUI07.png)

## Blob Storage.

You models and Data will be stored in a storage bucket attached to your namespace.  

Please remember to store Portected B classifid data and models in a Protected B bucket, and un-protected data and models in an unclassified bucket.

Please see the following link for more information on this topic:

[Azure Blob storage](https://statcan.github.io/aaw/en/5-Storage/AzureBlobStorage/)
