#include <stdio.h>
#include "pipeline.h"

char* pipeline_json = "{\"id\":\"c-demo\",\"processors\":[{\"lowercase\":{\"field\":\"user_name\"}}]}";

int main() {
    printf("Using go-event-pipeline from C:\n");

    // Initialize pipeline.
    printf("Loading pipeline: %s\n", pipeline_json);
    if (Load(pipeline_json)) {
        printf("Failed to load pipeline!\n");
        return 1;
    }

    // Process JSON with the pipeline.
    char *in = "{\"user_name\":\"John Doe\"}";
    printf("Calling Process() with input = %s\n", in);
    char* out = Process(in);
    printf("Process() returned: %s\n", out);
}
