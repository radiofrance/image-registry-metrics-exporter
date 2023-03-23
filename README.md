
# ImageRegistryMetrics

**Image Registry Metrics Exporter (IRME)** provide metrics about creation and uploading time of images on OCI Registry.

It currently supports only Google Container Registry.

Metrics are available on `localhost:8080/metrics`. Scheme is :

``image_registry_exporter_tag_build_time``
``image_registry_exporter_tag_uploaded_time``  
with labels  `image` and `tag`  .


## Local testing
- Create a file with basic configuration in `./config.yaml`, `$HOME/.irme/config.yaml` or `/etc/irme/config.yaml`:
````
registries:  
- provider: google      # Defines backend compatibility. Only google (Google) is available now.  
  domain: gcr.io     # Which Registry endpoint it should use  
  imagesFilters:        # You can filters which repositories or images it should watch. For everything, set : .*  
  - distroless  
  tagsFilters:          # you can filterw which tags it should generate metrics on. For everything, set : .*  
  - .*  
  rateLimitAPI: 5      # Set how many calls to backend API it should do each seconds.  
  maxConcurrentJobs: 5 # Set How many image are processed in parallel  
  cron: "0 * * * *" # After first execution, set a cron job to regenerate metrics.
````
- `make artifact` ;
- `./gobin` ;
- `curl localhost:8080/metrics` should output tag and images metrics.

## Configuration

### Google Authentication

IRME required a service account with following roles :
- `roles/storage.objectViewer` for repositories defined in `imagesFilters` ;
- `roles/browser` to be able to list all images available on repositories.

Without Helm, credentials should be exported by `GOOGLE_APPLICATION_CREDENTIALS` or configured on docker configuration file. For Helm implementation, check specific informations on Helm `values.yaml`.

### Configuration  file
Configuration is defined by this struct :
````  
registries:  
- provider: google      # Defines backend compatibility. Only google (Google) is available now.  
  domain: eu.gcr.io     # Which Registry endpoint it should use  
  imagesFilters:        # You can filters which repositories or images it should watch. For everything, set : .*  
  - filter  
  tagsFilters:          # you can filterw which tags it should generate metrics on. For everything, set : .*  
  - ^latest$  
  rateLimitAPI: 5      # Set how many calls to backend API it should do each seconds.   
  maxConcurrentJobs: 5 # Set How many image are processed in parallel  
  cron: "0 * * * *"    # After first execution, set a cron job to regenerate metrics.   
````    

- For Helm installation, configuration is available at key `configmap` with multilines following next convention
- Without Helm, configuration should be provided in a file exporter by environment variable `IRME_CONF_FILE_PATH`  
  
