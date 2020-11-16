//
// Copyright (c) 2012-2019 Red Hat, Inc.	
// This program and the accompanying materials are made	
// available under the terms of the Eclipse Public License 2.0	
// which is available at https://www.eclipse.org/legal/epl-2.0/	
//	
// SPDX-License-Identifier: EPL-2.0	
//	
// Contributors:	// Contributors:
//   Red Hat, Inc. - initial API and implementation	//   Red Hat, Inc. - initial API and implementation
//

package controller

import (
	"github.com/eclipse/che-operator/pkg/controller/che"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, che.Add)
}
