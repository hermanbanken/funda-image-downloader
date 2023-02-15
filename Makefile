.PHONY: submit
submit:
	gcloud builds submit . --config=./cloudbuild.yaml
