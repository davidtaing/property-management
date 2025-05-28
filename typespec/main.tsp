import "@typespec/http";

using Http;
@service(#{ title: "Property Management Service" })
@server("http://localhost:8080", "Local Development Server")
@server("https://property-management.fly.dev", "Production Server")

namespace PropertyManagementService;

model Address {
  address_line_1: string;
  address_line_2?: string;
  suburb: string;
  postcode: string;
  state: string;
  country: string;
}

model Landlord {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    name: string;
    @format("email")
    email: string;
    mobile: string;
    phone?: string;
    ...Address;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
}

model CreateLandlord {
  name: string;
  @format("email")
  email: string;
  mobile: string;
  phone?: string;
  ...Address;
}

model LandlordList {
  items: Landlord[];
  pagination: PaginatedMetadata;
}

model Property {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    @format("uuid")
    landlord_id: string;
    ...Address;
    management_fee: float64;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
    
}

model CreateProperty {
  @format("uuid")
  landlord_id: string;
  ...Address;
  management_fee: float64;
}

model PropertyList {
  items: Property[];
  pagination: PaginatedMetadata;
}

model Tenant {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    @format("uuid")
    property_id: string;
    name: string;
    @format("email")
    email: string;
    mobile: string;
    phone?: string;
    original_start_date: plainDate;
    start_date: plainDate;
    end_date: plainDate;
    vacate_date?: plainDate;
    termination_date?: plainDate;
    termination_reason?: string;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
} 

model CreateTenant {
  @format("uuid")
  property_id: string;
  name: string;
  @format("email")
  email: string;
  mobile: string;
  phone?: string;
  original_start_date: plainDate;
  start_date: plainDate;
  end_date: plainDate;
}

model TenantList {
  items: Tenant[];
  pagination: PaginatedMetadata;
}

@error
model Error {
  code: int32;
  message: string;
}

model PaginatedMetadata {
  total: int32;
  count: int32;
  per_page: int32;
  current_page: int32;
  total_pages: int32;
}

@route("/landlords")
namespace Landlords {
  @tag("Landlord")
  @get
  op list(@query page?: int32, @query limit?: int32, @query name?: string): {
    @statusCode statusCode: 200;
    @body landlords: LandlordList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error
  };

  @tag("Landlord")
  @get
  op get(@path id: string): {
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
  op create(@body landlord: CreateLandlord): {
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
  op update(@path id: string, @body landlord: Landlord): {
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
  op archive(@path id: string): {
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
namespace Properties {
  @tag("Property")
  @get
  op list(@query page?: int32, @query limit?: int32, @query address?: string): {
    @statusCode statusCode: 200;
    @body properties: PropertyList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Property")
  @get
  op get(@path id: string): {
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
  op create(@body property: CreateProperty): {
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
  op update(@path id: string, @body property: Property): {
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
  op archive(@path id: string): {
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
  op list(@query page?: int32, @query limit?: int32, @query name?: string): {
    @statusCode statusCode: 200;
    @body tenants: TenantList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @tag("Tenant")
  @get
  op get(@path id: string): {
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
  op create(@body tenant: CreateTenant): {
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
  op update(@path id: string, @body tenant: Tenant): {
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
  op archive(@path id: string): {
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