#!/bin/bash
# https://cn.vuejs.org/index.html
cd vueVisitor/ && yarn run build && rm -rf dist && mv public dist