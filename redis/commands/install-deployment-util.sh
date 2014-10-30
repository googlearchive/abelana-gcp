#!/bin/bash

# Copyright 2014 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# deployment_util.sh
#
# Utility script for deployments under Deployment Manager infrastructure.
# This script writes its contents to a well-known location for other scripts
# to source as the Deployment Manager infrastructure does not have a
# direct way to upload and then reference such scripts.
#

mkdir -p ${DEPLOY_INSTALL_TEMP}

# Quote the limit string (EOF) to suppress parameter substitution
cat > ${DEPLOY_INSTALL_TEMP}/deployment_util.sh << 'EOF'

# Copyright 2014 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# get_instance_list
#
# Returns a list of the instances (newline-delimited)
# from the specified resource view.
function get_instance_list {
  local view_name="$1"
  local zone="$2"

  gcloud preview resource-views --zone=${zone} \
    resources --resourceview=${view_name} list -l
}
readonly -f get_instance_list

EOF
