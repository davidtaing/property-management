import "@typespec/http";
import "./common.tsp";

using Http;
@service(#{ title: "Property Management Service" })
@server("http://localhost:8080", "Local Development Server")

namespace PropertyManagementService;

using PropertyManagementService.Types;

model Landlord {
    @visibility(Lifecycle.Read)
    id: id;
    name: string;
    email: string;
    mobile: string;
    phone: string;
    is_archived?: offsetDateTime;
}

model Property {
    @visibility(Lifecycle.Read)
    id: id;
    landlord_id: id;
    address_line_1: string;
    address_line_2?: string;
    suburb: string;
    postcode: string;
    state: string;
    country: string;
    management_fee: decimal;
    is_archived?: offsetDateTime;
}

model Tenant {
    @visibility(Lifecycle.Read)
    id: id;
    property_id: id;
    name: string;
    email: string;
    mobile: string;
    phone: string;
    original_start_date: offsetDateTime;
    start_date: offsetDateTime;
    end_date: offsetDateTime;
    vacate_date?: plainDate;
    termination_date?: plainDate;
    termination_reason?: string;
    is_archived?: offsetDateTime;
} 

@error
model Error {
  code: int32;
  message: string;
}

@route("/landlords")
namespace Landlords {
  @tag("Landlord")
  @get
  op list(): {
    @statusCode statusCode: 200;
    @body landlords: Landlord[]
  } | {
    @statusCode statusCode: 500;
    @body error: Error
  };

  @tag("Landlord")
  @get
  op get(@path id: id): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Landlord")
  @post
  op create(@body landlord: Landlord): {
    @statusCode statusCode: 201;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Landlord")
  @put
  op update(@path id: id, @body landlord: Landlord): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Landlord")
  @delete
  op archive(@path id: id): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}

@route("/properties")
namespace Proeprties {
  @tag("Property")
  @get
  op list(): {
    @statusCode statusCode: 200;
    @body properties: Property[]
  } | {
    @statusCode statusCode: 500;
    @body error: Error
  };

  @tag("Property")
  @get
  op get(@path id: id): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Property")
  @post
  op create(@body property: Property): {
    @statusCode statusCode: 201;
    @body property: Property;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Property")
  @put
  op update(@path id: id, @body property: Property): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Property")
  @delete
  op archive(@path id: id): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}

@route("/tenants")
namespace Tenants {
  @tag("Tenant")
  @get
  op list(): {
    @statusCode statusCode: 200;
    @body tenants: Tenant[]
  } | {
    @statusCode statusCode: 500;
    @body error: Error
  };

  @tag("Tenant")
  @get
  op get(@path id: id): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Tenant")
  @post
  op create(@body tenant: Tenant): {
    @statusCode statusCode: 201;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Tenant")
  @put
  op update(@path id: id, @body tenant: Tenant): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Tenant")
  @delete
  op archive(@path id: id): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}