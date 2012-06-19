# Copyright <year> <name>. All rights reserved.
# Use of this source code is governed by license 
# that can be found in the LICENSE file.

app: 
   cmd: {{ .name }}server
   pkg: {{ .name }}

server:
  port: {{ .port }}
  cacheTemplates: false

database:
  driver: {{ .driver }}
  connect: {{ .connect }}

encryption:
  salt: {{ .salt }}
  key: {{ .key }}
